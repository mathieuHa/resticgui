package secrets

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBox(t *testing.T, fill byte) *Box {
	t.Helper()
	b, err := New(bytes.Repeat([]byte{fill}, KeySize))
	require.NoError(t, err)
	return b
}

func TestSealOpenRoundTrip(t *testing.T) {
	b := newBox(t, 1)
	for _, plaintext := range [][]byte{{}, []byte("hunter2"), bytes.Repeat([]byte("x"), 1<<16)} {
		sealed, err := b.Seal(plaintext)
		require.NoError(t, err)
		assert.False(t, len(plaintext) > 0 && bytes.Contains(sealed, plaintext), "plaintext visible in sealed value")
		opened, err := b.Open(sealed)
		require.NoError(t, err)
		assert.Equal(t, plaintext, append([]byte{}, opened...))
	}
}

func TestSealUsesFreshNonce(t *testing.T) {
	b := newBox(t, 1)
	first, err := b.Seal([]byte("same"))
	require.NoError(t, err)
	second, err := b.Seal([]byte("same"))
	require.NoError(t, err)
	assert.NotEqual(t, first, second)
}

func TestOpenRejects(t *testing.T) {
	b := newBox(t, 1)
	sealed, err := b.Seal([]byte("secret"))
	require.NoError(t, err)

	t.Run("wrong key", func(t *testing.T) {
		_, err := newBox(t, 2).Open(sealed)
		assert.ErrorIs(t, err, ErrDecrypt)
	})
	t.Run("tampered ciphertext", func(t *testing.T) {
		tampered := append([]byte{}, sealed...)
		tampered[len(tampered)-1] ^= 1
		_, err := b.Open(tampered)
		assert.ErrorIs(t, err, ErrDecrypt)
	})
	t.Run("unknown version", func(t *testing.T) {
		tampered := append([]byte{}, sealed...)
		tampered[0] = 9
		_, err := b.Open(tampered)
		assert.ErrorIs(t, err, ErrDecrypt)
	})
	t.Run("truncated", func(t *testing.T) {
		_, err := b.Open(sealed[:5])
		assert.ErrorIs(t, err, ErrDecrypt)
	})
}

func TestNewRejectsBadKeySize(t *testing.T) {
	_, err := New(make([]byte, 16))
	assert.Error(t, err)
}

func TestLoadKey(t *testing.T) {
	t.Run("creates a key file once and reuses it", func(t *testing.T) {
		t.Setenv(KeyEnv, "")
		os.Unsetenv(KeyEnv)
		path := filepath.Join(t.TempDir(), "resticgui.key")

		first, err := LoadKey(path)
		require.NoError(t, err)
		assert.Len(t, first, KeySize)
		if runtime.GOOS != "windows" {
			info, err := os.Stat(path)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
		}

		second, err := LoadKey(path)
		require.NoError(t, err)
		assert.Equal(t, first, second)
	})

	t.Run("environment wins over the file", func(t *testing.T) {
		key := bytes.Repeat([]byte{7}, KeySize)
		t.Setenv(KeyEnv, base64.StdEncoding.EncodeToString(key))
		got, err := LoadKey(filepath.Join(t.TempDir(), "unused.key"))
		require.NoError(t, err)
		assert.Equal(t, key, got)
	})

	t.Run("invalid environment key", func(t *testing.T) {
		t.Setenv(KeyEnv, base64.StdEncoding.EncodeToString([]byte("short")))
		_, err := LoadKey(filepath.Join(t.TempDir(), "unused.key"))
		assert.Error(t, err)
	})

	t.Run("refuses a world readable key file", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("permissions are not enforced on windows")
		}
		t.Setenv(KeyEnv, "")
		os.Unsetenv(KeyEnv)
		path := filepath.Join(t.TempDir(), "open.key")
		require.NoError(t, os.WriteFile(path, []byte(base64.StdEncoding.EncodeToString(make([]byte, KeySize))), 0o644))
		_, err := LoadKey(path)
		assert.ErrorContains(t, err, "chmod 600")
	})
}
