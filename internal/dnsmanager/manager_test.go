package dnsmanager

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "resolv.conf")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	initial := `nameserver 8.8.8.8
nameserver 1.1.1.1
search example.com
options rotate`
	_, err = tmpFile.WriteString(initial)
	require.NoError(t, err)
	tmpFile.Close()

	m := NewManager(tmpFile.Name())

	t.Run("GetAll", func(t *testing.T) {
		cfg, err := m.GetAll()
		require.NoError(t, err)
		assert.Contains(t, cfg.Nameservers, "8.8.8.8")
		assert.Contains(t, cfg.Nameservers, "1.1.1.1")
	})

	t.Run("Add success", func(t *testing.T) {
		err := m.Add("8.8.4.4")
		require.NoError(t, err)

		cfg, _ := m.GetAll()
		assert.Contains(t, cfg.Nameservers, "8.8.4.4")
	})

	t.Run("Add duplicate", func(t *testing.T) {
		err := m.Add("8.8.4.4")
		assert.ErrorIs(t, err, ErrAlreadyExists)
	})

	t.Run("Remove success", func(t *testing.T) {
		err := m.Remove("1.1.1.1")
		require.NoError(t, err)

		cfg, _ := m.GetAll()
		assert.NotContains(t, cfg.Nameservers, "1.1.1.1")
	})

	t.Run("Remove not found", func(t *testing.T) {
		err := m.Remove("9.9.9.9")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("Invalid IP", func(t *testing.T) {
		err := m.Add("999.999.999.999")
		assert.ErrorIs(t, err, ErrInvalidIP)

		err = m.Remove("invalid-ip")
		assert.ErrorIs(t, err, ErrInvalidIP)
	})

	t.Run("Empty file", func(t *testing.T) {
		emptyFile, err := os.CreateTemp("", "empty-resolv.conf")
		require.NoError(t, err)
		defer os.Remove(emptyFile.Name())
		emptyFile.Close()

		emptyManager := NewManager(emptyFile.Name())
		cfg, err := emptyManager.GetAll()
		require.NoError(t, err)
		assert.Empty(t, cfg.Nameservers)
	})
}
