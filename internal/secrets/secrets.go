// Package secrets encrypts small values at rest (repository passwords and
// backend credentials) with AES-256-GCM.
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// KeySize is the length in bytes of an AES-256 key.
const KeySize = 32

// KeyEnv is the environment variable holding a base64 key. It takes
// precedence over the key file.
const KeyEnv = "RESTICGUI_SECRET_KEY"

// version prefixes every sealed value so the format can change later.
const version byte = 1

var additionalData = []byte("resticgui/secrets/v1")

// ErrDecrypt is returned when a value cannot be opened, either because the
// key is wrong or because the data was altered.
var ErrDecrypt = errors.New("secrets: cannot decrypt value (wrong key or corrupted data)")

// Box seals and opens values with a single key.
type Box struct {
	aead cipher.AEAD
}

// New returns a Box using a KeySize-byte key.
func New(key []byte) (*Box, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("secrets: key must be %d bytes, got %d", KeySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Box{aead: aead}, nil
}

// Seal encrypts plaintext. The output is version || nonce || ciphertext.
func (b *Box) Seal(plaintext []byte) ([]byte, error) {
	out := make([]byte, 1+b.aead.NonceSize(), 1+b.aead.NonceSize()+len(plaintext)+b.aead.Overhead())
	out[0] = version
	if _, err := rand.Read(out[1:]); err != nil {
		return nil, err
	}
	return b.aead.Seal(out, out[1:], plaintext, additionalData), nil
}

// Open decrypts a value produced by Seal.
func (b *Box) Open(sealed []byte) ([]byte, error) {
	nonceSize := b.aead.NonceSize()
	if len(sealed) < 1+nonceSize+b.aead.Overhead() || sealed[0] != version {
		return nil, ErrDecrypt
	}
	plaintext, err := b.aead.Open(nil, sealed[1:1+nonceSize], sealed[1+nonceSize:], additionalData)
	if err != nil {
		return nil, ErrDecrypt
	}
	return plaintext, nil
}

// LoadKey returns the key from $RESTICGUI_SECRET_KEY if set, otherwise from
// keyFile. When neither exists, a new random key is written to keyFile with
// 0600 permissions. A key file readable by group or others is refused.
func LoadKey(keyFile string) ([]byte, error) {
	if encoded, ok := os.LookupEnv(KeyEnv); ok {
		key, err := decodeKey(encoded)
		if err != nil {
			return nil, fmt.Errorf("secrets: %s: %w", KeyEnv, err)
		}
		return key, nil
	}

	info, err := os.Stat(keyFile)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return createKeyFile(keyFile)
	case err != nil:
		return nil, err
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("secrets: key file %s is accessible by other users, run: chmod 600 %s", keyFile, keyFile)
	}
	content, err := os.ReadFile(keyFile) // #nosec G304 -- the key file named in the configuration
	if err != nil {
		return nil, err
	}
	key, err := decodeKey(string(content))
	if err != nil {
		return nil, fmt.Errorf("secrets: key file %s: %w", keyFile, err)
	}
	return key, nil
}

func createKeyFile(keyFile string) ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	// O_EXCL so two processes starting together cannot overwrite each other's key.
	f, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- the key file named in the configuration
	if err != nil {
		return nil, err
	}
	if _, err := f.WriteString(base64.StdEncoding.EncodeToString(key) + "\n"); err != nil {
		f.Close()
		return nil, err
	}
	return key, f.Close()
}

func decodeKey(encoded string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, errors.New("key is not valid base64")
	}
	if len(key) != KeySize {
		return nil, fmt.Errorf("key must decode to %d bytes, got %d", KeySize, len(key))
	}
	return key, nil
}
