package share

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func Run(args []string, version string, input io.Reader, output io.Writer) error {
	flags := flag.NewFlagSet("pc-to-iphone-share", flag.ContinueOnError)
	flags.SetOutput(output)
	folder := flags.String("folder", "", "folder to share (interactive when omitted)")
	bind := flags.String("bind", "", "private IPv4 interface to use (select interactively when omitted)")
	port := flags.Int("port", 9443, "HTTPS port; never modifies firewall or router settings")
	state := flags.String("state", "", "private settings directory (defaults outside the project)")
	reset := flags.Bool("reset-password", false, "generate a new password and invalidate sessions on restart")
	showVersion := flags.Bool("version", false, "print the version and exit")
	loopback := flags.Bool("loopback-only", false, "isolated local testing; phone access is disabled")
	check := flags.Bool("check", false, "validate folder, settings and network without starting servers")
	shareFor := flags.Duration("share-for", 0, "automatically stop after a duration, e.g. 30m (maximum 24h)")
	maxDownloads := flags.Int("max-downloads", 4, "maximum concurrent file/ZIP/checksum/speed operations (1-16)")
	var excludes stringList
	flags.Var(&excludes, "exclude", "case-insensitive filename glob to hide at every depth; repeat as needed")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected arguments; use --help")
	}
	if *showVersion {
		fmt.Fprintln(output, "PC to iPhone Share", version)
		return nil
	}
	options := Options{MaxDownloads: *maxDownloads, ShareFor: *shareFor, Exclude: excludes}
	if err := options.validate(); err != nil {
		return err
	}
	if *state == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return errors.New("cannot find an OS private settings directory")
		}
		*state = filepath.Join(dir, "PCToIPhoneShare")
	}
	var err error
	*state, err = filepath.Abs(*state)
	if err != nil {
		return errors.New("invalid private settings directory")
	}
	if info, err := os.Lstat(*state); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("private settings must not be a linked directory")
	}
	if err = os.MkdirAll(*state, 0700); err != nil {
		return errors.New("cannot create private settings directory")
	}
	if err = protectState(*state); err != nil {
		return err
	}
	config, err := loadConfig(*state)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(input)
	prompt := func(label string) (string, error) {
		fmt.Fprint(output, label)
		value, err := reader.ReadString('\n')
		if err != nil && len(value) == 0 {
			return "", errors.New("interactive input unavailable; pass --folder and --bind")
		}
		return strings.Trim(strings.TrimSpace(value), "\"'"), nil
	}
	if *folder == "" {
		label := "Folder to share (paste or drag its path): "
		if config.Folder != "" {
			label = "Folder to share (Enter keeps the previous folder): "
		}
		value, err := prompt(label)
		if err != nil {
			return err
		}
		if value == "" {
			*folder = config.Folder
		} else {
			*folder = value
		}
	}
	if *folder == "" {
		return errors.New("choose one folder to share")
	}
	*folder, err = filepath.Abs(*folder)
	if err != nil {
		return errors.New("invalid folder path")
	}
	var selected Network
	if *loopback {
		if *bind != "" && *bind != "127.0.0.1" {
			return errors.New("loopback testing cannot bind a network address")
		}
		selected = Network{Address: "127.0.0.1", Prefix: 32}
	}
	if !*loopback && *bind != "" {
		selected, err = networkFor(*bind)
		if err != nil {
			return err
		}
	}
	if !*loopback && *bind == "" {
		candidates := networks()
		if len(candidates) == 0 {
			return errors.New("no active private IPv4 network found; connect to Wi-Fi first")
		}
		for i, n := range candidates {
			fmt.Fprintf(output, "%d. %s — %s/%d\n", i+1, n.Name, n.Address, n.Prefix)
		}
		if len(candidates) == 1 {
			selected = candidates[0]
		} else {
			value, err := prompt("Select your Wi-Fi/LAN interface number (avoid VPN/virtual adapters): ")
			if err != nil {
				return err
			}
			index, err := strconv.Atoi(value)
			if err != nil || index < 1 || index > len(candidates) {
				return errors.New("invalid network selection")
			}
			selected = candidates[index-1]
		}
	}
	config.Folder, config.Bind, config.Prefix, config.Port = *folder, selected.Address, selected.Prefix, *port
	password := ""
	if config.PasswordHash == "" || *reset {
		password, err = newPassword(&config)
		if err != nil {
			return errors.New("cannot generate password")
		}
	}
	if err = validateConfig(config, *state, *loopback); err != nil {
		return err
	}
	if *check {
		f, err := os.Open(config.Folder)
		if err != nil {
			return errors.New("folder is not readable")
		}
		defer f.Close()
		if _, err = f.Readdirnames(1); err != nil && err != io.EOF {
			return errors.New("folder is not readable")
		}
		fmt.Fprintln(output, "Setup checks passed: folder readable, private settings separated, selected interface valid. No servers started. Firewall reachability and phone trust still require a connection test.")
		return nil
	}
	certificates, err := ensureCertificates(*state, config.Bind)
	if err != nil {
		return err
	}
	app, err := NewAppWithOptions(config, certificates, options)
	if err != nil {
		return err
	}
	defer app.Close()
	listener, err := net.Listen("tcp4", app.Address())
	if err != nil {
		return errors.New("HTTPS port is unavailable; choose another --port (existing servers are never stopped)")
	}
	defer listener.Close()
	localListener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return errors.New("cannot start the local dashboard")
	}
	defer localListener.Close()
	if err = saveConfig(*state, config); err != nil {
		return err
	}
	stop := make(chan struct{})
	dashboard := &Dashboard{App: app, Password: password, Token: randomToken(32), Host: localListener.Addr().String(), Version: version, Stop: stop}
	server := httpServer(app)
	server.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{certificates.TLS}}
	localServer := httpServer(dashboard)
	failures := make(chan error, 2)
	go func() {
		failures <- server.ServeTLS(&cappedListener{Listener: listener, slots: make(chan struct{}, 64)}, "", "")
	}()
	go func() { failures <- localServer.Serve(localListener) }()
	fmt.Fprintf(output, "\nPC to iPhone Share %s\nPhone URL: %s\nLocal dashboard: http://%s/#%s\n", version, app.Origin(), dashboard.Host, dashboard.Token)
	if password != "" {
		fmt.Fprintln(output, "New share password:", password)
	} else {
		fmt.Fprintln(output, "Existing password unchanged. Use --reset-password if you lost it.")
	}
	fmt.Fprintln(output, "Certificate SHA-256:", certificates.Fingerprint)
	fmt.Fprintln(output, "Keep this window open. Press Ctrl+C to stop sharing. No public hosting, router changes, or telemetry.")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	var expiry <-chan time.Time
	if *shareFor > 0 {
		timer := time.NewTimer(time.Until(app.expires))
		defer timer.Stop()
		expiry = timer.C
	}
	select {
	case <-ctx.Done():
	case <-stop:
	case <-expiry:
		fmt.Fprintln(output, "Sharing window expired; stopping new access.")
	case err := <-failures:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return errors.New("a local server stopped unexpectedly")
		}
	}
	shutdown, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelShutdown()
	// Existing streams can finish during shutdown. After 30 seconds, remaining connections close.
	_ = localServer.Shutdown(shutdown)
	_ = server.Shutdown(shutdown)
	_ = localServer.Close()
	_ = server.Close()
	fmt.Fprintln(output, "Sharing stopped.")
	return nil
}

type stringList []string

func (s *stringList) String() string         { return strings.Join(*s, ", ") }
func (s *stringList) Set(value string) error { *s = append(*s, value); return nil }
