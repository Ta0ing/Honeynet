package agentupdate

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const signingKeyFile = "update-signing.key"

var versionPattern = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z._+-]{0,63}$`)

type Descriptor struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	SHA256  string `json:"sha256"`
	Size    int64  `json:"size"`
}

type Signer struct {
	private ed25519.PrivateKey
	public  ed25519.PublicKey
}

func LoadOrCreateSigner(dir string) (*Signer, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, signingKeyFile)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		_, private, keyErr := ed25519.GenerateKey(rand.Reader)
		if keyErr != nil {
			return nil, keyErr
		}
		der, keyErr := x509.MarshalPKCS8PrivateKey(private)
		if keyErr != nil {
			return nil, keyErr
		}
		data = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
		if keyErr := writeAtomic(path, data, 0600); keyErr != nil {
			return nil, keyErr
		}
	} else if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0600); err != nil {
		return nil, fmt.Errorf("secure update signing key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid update signing key PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse update signing key: %w", err)
	}
	private, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("update signing key is not Ed25519")
	}
	public := private.Public().(ed25519.PublicKey)
	return &Signer{private: private, public: public}, nil
}

func (s *Signer) Sign(descriptor Descriptor) (string, error) {
	message, err := Canonical(descriptor)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(ed25519.Sign(s.private, message)), nil
}

func (s *Signer) PublicKey() string {
	return base64.RawStdEncoding.EncodeToString(s.public)
}

func (s *Signer) KeyID() string {
	sum := sha256.Sum256(s.public)
	return hex.EncodeToString(sum[:8])
}

func Verify(publicKey, signature string, descriptor Descriptor) error {
	public, err := base64.RawStdEncoding.DecodeString(strings.TrimSpace(publicKey))
	if err != nil || len(public) != ed25519.PublicKeySize {
		return errors.New("invalid update public key")
	}
	signed, err := base64.RawStdEncoding.DecodeString(strings.TrimSpace(signature))
	if err != nil || len(signed) != ed25519.SignatureSize {
		return errors.New("invalid update signature")
	}
	message, err := Canonical(descriptor)
	if err != nil {
		return err
	}
	if !ed25519.Verify(ed25519.PublicKey(public), message, signed) {
		return errors.New("Agent update signature verification failed")
	}
	return nil
}

func Canonical(descriptor Descriptor) ([]byte, error) {
	descriptor.Version = strings.TrimSpace(descriptor.Version)
	descriptor.OS = strings.ToLower(strings.TrimSpace(descriptor.OS))
	descriptor.Arch = strings.ToLower(strings.TrimSpace(descriptor.Arch))
	descriptor.SHA256 = strings.ToLower(strings.TrimSpace(descriptor.SHA256))
	if !versionPattern.MatchString(descriptor.Version) {
		return nil, errors.New("invalid update version")
	}
	if descriptor.OS == "" || descriptor.Arch == "" || len(descriptor.SHA256) != 64 || descriptor.Size <= 0 {
		return nil, errors.New("incomplete update descriptor")
	}
	if _, err := hex.DecodeString(descriptor.SHA256); err != nil {
		return nil, errors.New("invalid update SHA-256")
	}
	return []byte(fmt.Sprintf("honeynet-agent-update-v1\n%s\n%s\n%s\n%s\n%d\n", descriptor.Version, descriptor.OS, descriptor.Arch, descriptor.SHA256, descriptor.Size)), nil
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, mode); err != nil {
		return err
	}
	if err := os.Chmod(temporary, mode); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
