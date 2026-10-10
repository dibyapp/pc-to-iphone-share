package share

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Certificates struct {
	TLS          tls.Certificate
	DER, Profile []byte
	Fingerprint  string
}

func serialNumber() *big.Int {
	n, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		panic("operating system randomness unavailable")
	}
	return n
}

func ensureCertificates(dir, address string) (Certificates, error) {
	var result Certificates
	caFile, keyFile := filepath.Join(dir, "ca.pem"), filepath.Join(dir, "ca.key")
	caPEM, readErr := os.ReadFile(caFile)
	var ca *x509.Certificate
	var caKey *ecdsa.PrivateKey
	if readErr == nil {
		block, _ := pem.Decode(caPEM)
		if block == nil {
			return result, errors.New("invalid local certificate authority")
		}
		var err error
		ca, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return result, errors.New("invalid local certificate authority")
		}
		keyPEM, err := os.ReadFile(keyFile)
		if err != nil {
			return result, errors.New("cannot read local certificate key")
		}
		keyBlock, _ := pem.Decode(keyPEM)
		if keyBlock == nil {
			return result, errors.New("invalid local certificate key")
		}
		parsed, err := x509.ParseECPrivateKey(keyBlock.Bytes)
		if err != nil {
			return result, errors.New("invalid local certificate key")
		}
		caKey = parsed
		publicKey, ok := ca.PublicKey.(*ecdsa.PublicKey)
		if !ok || !publicKey.Equal(&caKey.PublicKey) || ca.CheckSignatureFrom(ca) != nil {
			return result, errors.New("local certificate authority and key do not match")
		}
		if !ca.IsCA || time.Until(ca.NotAfter) < 24*time.Hour {
			return result, errors.New("local certificate authority expired; use a new private settings directory and re-pair your devices")
		}
	} else if os.IsNotExist(readErr) {
		var err error
		caKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return result, err
		}
		template := &x509.Certificate{SerialNumber: serialNumber(), Subject: pkix.Name{CommonName: "PC to iPhone Share Local Certificate"},
			NotBefore: time.Now().Add(-24 * time.Hour), NotAfter: time.Now().AddDate(5, 0, 0), IsCA: true,
			BasicConstraintsValid: true, MaxPathLen: 0, MaxPathLenZero: true,
			KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
			// Constrain this authority to private IPv4 networks; disallow real DNS names.
			PermittedDNSDomainsCritical: true, PermittedDNSDomains: []string{".pc-to-iphone-share.invalid"},
			PermittedIPRanges: []*net.IPNet{{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(8, 32)}, {IP: net.ParseIP("172.16.0.0").To4(), Mask: net.CIDRMask(12, 32)}, {IP: net.ParseIP("192.168.0.0").To4(), Mask: net.CIDRMask(16, 32)}, {IP: net.ParseIP("127.0.0.1").To4(), Mask: net.CIDRMask(32, 32)}},
		}
		der, err := x509.CreateCertificate(rand.Reader, template, template, &caKey.PublicKey, caKey)
		if err != nil {
			return result, err
		}
		ca, err = x509.ParseCertificate(der)
		if err != nil {
			return result, err
		}
		caPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
		keyDER, err := x509.MarshalECPrivateKey(caKey)
		if err != nil {
			return result, err
		}
		if err = privateWrite(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})); err != nil {
			return result, err
		}
		if err = privateWrite(caFile, caPEM); err != nil {
			return result, err
		}
	} else {
		return result, errors.New("cannot read local certificate authority")
	}
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return result, err
	}
	template := &x509.Certificate{SerialNumber: serialNumber(), Subject: pkix.Name{CommonName: "PC to iPhone Share"},
		NotBefore: time.Now().Add(-24 * time.Hour), NotAfter: time.Now().AddDate(0, 0, 90),
		IPAddresses: []net.IP{net.ParseIP(address)}, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	leaf, err := x509.CreateCertificate(rand.Reader, template, ca, &leafKey.PublicKey, caKey)
	if err != nil {
		return result, err
	}
	leafKeyDER, err := x509.MarshalECPrivateKey(leafKey)
	if err != nil {
		return result, err
	}
	leafPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leaf})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: leafKeyDER})
	result.TLS, err = tls.X509KeyPair(leafPEM, keyPEM)
	if err != nil {
		return result, err
	}
	result.DER = ca.Raw
	fingerprint := sha256.Sum256(ca.Raw)
	result.Fingerprint = hex.EncodeToString(fingerprint[:])
	// Only public certificate data goes into the profile. Each installation has a unique identifier.
	id := "local.pc.to.iphone.share." + result.Fingerprint[:24]
	profile := `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"><plist version="1.0"><dict><key>PayloadType</key><string>Configuration</string><key>PayloadVersion</key><integer>1</integer><key>PayloadIdentifier</key><string>%s</string><key>PayloadUUID</key><string>%s</string><key>PayloadDisplayName</key><string>PC to iPhone Share Local Certificate</string><key>PayloadDescription</key><string>One public certificate for your local file-sharing app. No VPN or device management.</string><key>PayloadContent</key><array><dict><key>PayloadType</key><string>com.apple.security.root</string><key>PayloadVersion</key><integer>1</integer><key>PayloadIdentifier</key><string>%s.certificate</string><key>PayloadUUID</key><string>%s</string><key>PayloadDisplayName</key><string>PC to iPhone Share Local Certificate</string><key>PayloadCertificateFileName</key><string>PC-to-iPhone-Share.cer</string><key>PayloadContent</key><data>%s</data></dict></array></dict></plist>`
	result.Profile = []byte(fmt.Sprintf(profile, id, uuid(), id, uuid(), base64.StdEncoding.EncodeToString(ca.Raw)))
	if err = privateWrite(filepath.Join(dir, "pc-to-iphone-share.cer"), ca.Raw); err != nil {
		return result, err
	}
	if err = privateWrite(filepath.Join(dir, "iphone.mobileconfig"), result.Profile); err != nil {
		return result, err
	}
	return result, nil
}

func uuid() string {
	b := randomBytes(16)
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}
