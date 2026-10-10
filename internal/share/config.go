package share

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const passwordIterations = 600000

type Config struct {
	Folder       string `json:"folder"`
	Bind         string `json:"bind"`
	Prefix       int    `json:"prefix"`
	Port         int    `json:"port"`
	Salt         string `json:"salt"`
	PasswordHash string `json:"password_hash"`
	Iterations   int    `json:"iterations"`
}

func randomToken(n int) string { return base64.RawURLEncoding.EncodeToString(randomBytes(n)) }
func randomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("operating system randomness unavailable")
	}
	return b
}

func derive(password string, salt []byte, iterations int) ([]byte, error) {
	return pbkdf2.Key(sha256.New, password, salt, iterations, 32)
}

func newPassword(c *Config) (string, error) {
	password := randomToken(24)
	salt := randomBytes(32)
	key, err := derive(password, salt, passwordIterations)
	if err != nil {
		return "", err
	}
	c.Salt, c.PasswordHash, c.Iterations = hex.EncodeToString(salt), hex.EncodeToString(key), passwordIterations
	return password, nil
}

func loadConfig(dir string) (Config, error) {
	var c Config
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return c, errors.New("cannot read private configuration")
	}
	if err = json.Unmarshal(data, &c); err != nil {
		return c, errors.New("private configuration is invalid")
	}
	return c, nil
}

func saveConfig(dir string, c Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return privateWrite(filepath.Join(dir, "config.json"), data)
}

func privateWrite(name string, data []byte) error {
	if info, err := os.Lstat(name); err == nil && (!info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0) {
		return errors.New("private settings must be ordinary files, not links")
	} else if err != nil && !os.IsNotExist(err) {
		return errors.New("cannot inspect private settings")
	}
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.New("cannot write private settings")
	}
	defer f.Close()
	if err := f.Chmod(0600); err != nil {
		return errors.New("cannot protect private settings")
	}
	if _, err = f.Write(data); err != nil {
		return errors.New("cannot write private settings")
	}
	return f.Sync()
}

func containsPath(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	return err == nil && (rel == "." || (!filepath.IsAbs(rel) && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))))
}

func validateConfig(c Config, stateDir string, allowLoopback bool) error {
	ip := net.ParseIP(c.Bind)
	if ip == nil || ip.To4() == nil || !(ip.IsPrivate() || (allowLoopback && ip.IsLoopback())) {
		return errors.New("choose a private IPv4 address on the network you want to share with")
	}
	if c.Prefix < 8 || c.Prefix > 32 || c.Port < 1024 || c.Port > 65535 {
		return errors.New("invalid subnet or port (use an unprivileged port from 1024 to 65535)")
	}
	if c.Iterations < passwordIterations || c.Iterations > 2000000 {
		return errors.New("invalid password hash settings")
	}
	salt, se := hex.DecodeString(c.Salt)
	key, ke := hex.DecodeString(c.PasswordHash)
	if se != nil || ke != nil || len(salt) != 32 || len(key) != 32 {
		return errors.New("invalid password hash settings")
	}
	info, err := os.Lstat(c.Folder)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("shared folder must be a real directory, not a linked folder")
	}
	real, err := filepath.EvalSymlinks(c.Folder)
	if err != nil {
		return errors.New("cannot resolve shared folder")
	}
	state, err := filepath.EvalSymlinks(stateDir)
	if err != nil || containsPath(real, state) {
		return errors.New("private settings must be outside the shared folder")
	}
	exe, err := os.Executable()
	if err == nil && containsPath(real, exe) {
		return errors.New("keep the app executable outside the shared folder")
	}
	return nil
}

type Network struct {
	Name, Address string
	Prefix        int
}

func networks() []Network {
	var result []Network
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ip, network, err := net.ParseCIDR(addr.String())
			if err != nil || ip.To4() == nil || !ip.IsPrivate() {
				continue
			}
			prefix, _ := network.Mask.Size()
			result = append(result, Network{iface.Name, ip.String(), prefix})
		}
	}
	return result
}

func networkFor(address string) (Network, error) {
	for _, n := range networks() {
		if n.Address == address {
			return n, nil
		}
	}
	return Network{}, fmt.Errorf("the selected address is not assigned to an active private network interface")
}
