package gateway

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DeviceIdentity holds the device keypair loaded from device.json.
type DeviceIdentity struct {
	DeviceID   string
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// deviceFile is the on-disk JSON format.
type deviceFile struct {
	Version       int    `json:"version"`
	DeviceID      string `json:"deviceId"`
	PublicKeyPem  string `json:"publicKeyPem"`
	PrivateKeyPem string `json:"privateKeyPem"`
}

// DefaultIdentityPath returns ~/.openclaw/identity/device.json.
func DefaultIdentityPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".openclaw", "identity", "device.json")
}

// LoadDeviceIdentity reads and parses a device identity JSON file.
func LoadDeviceIdentity(path string) (*DeviceIdentity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read device identity: %w", err)
	}

	var df deviceFile
	if err := json.Unmarshal(data, &df); err != nil {
		return nil, fmt.Errorf("parse device identity: %w", err)
	}

	privKey, err := parseEd25519PrivateKeyPEM(df.PrivateKeyPem)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	pubKey, err := extractRawPublicKey(df.PublicKeyPem)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &DeviceIdentity{
		DeviceID:   df.DeviceID,
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}, nil
}

// SignConnectPayload creates and signs the device connect payload using v3 pipe-delimited format.
// Format: v3|deviceId|clientId|clientMode|role|scopes|signedAtMs|token|nonce|platform|deviceFamily
func (d *DeviceIdentity) SignConnectPayload(nonce string, token string) (DeviceInfo, error) {
	now := time.Now().UnixMilli()

	tokenStr := ""
	if token != "" {
		tokenStr = token
	}

	platform := strings.ToLower(runtime.GOOS)
	deviceFamily := "" // empty for non-mobile

	payload := strings.Join([]string{
		"v3",
		d.DeviceID,
		"gateway-client",  // clientId
		"backend",         // clientMode
		"operator",        // role
		"operator.admin",  // scopes (comma-joined)
		fmt.Sprintf("%d", now), // signedAtMs
		tokenStr,          // token (empty string if none)
		nonce,             // nonce from challenge
		platform,          // platform
		deviceFamily,      // deviceFamily (empty)
	}, "|")

	sig := ed25519.Sign(d.PrivateKey, []byte(payload))

	return DeviceInfo{
		ID:        d.DeviceID,
		PublicKey: base64UrlEncode([]byte(d.PublicKey)),
		Signature: base64UrlEncode(sig),
		SignedAt:  now,
		Nonce:     nonce,
	}, nil
}

// extractRawPublicKey parses a PEM-encoded public key and returns the raw 32-byte Ed25519 key.
func extractRawPublicKey(pemStr string) (ed25519.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKIX: %w", err)
	}

	edPub, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an Ed25519 public key")
	}

	return edPub, nil
}

// parseEd25519PrivateKeyPEM parses a PEM-encoded private key.
func parseEd25519PrivateKeyPEM(pemStr string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS8: %w", err)
	}

	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an Ed25519 private key")
	}

	return edKey, nil
}

// base64UrlEncode encodes bytes as base64url without padding.
func base64UrlEncode(data []byte) string {
	s := base64.StdEncoding.EncodeToString(data)
	s = strings.ReplaceAll(s, "+", "-")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.TrimRight(s, "=")
	return s
}
