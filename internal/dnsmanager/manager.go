package dnsmanager

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/foggyeclipse/dns-manager-go/pkg/validator"
	"golang.org/x/sys/unix"
)

var (
	ErrAlreadyExists = errors.New("nameserver already exists")
	ErrNotFound      = errors.New("nameserver not found")
	ErrInvalidIP     = errors.New("invalid IP address")
)

type Manager struct {
	filePath string
	mu       sync.Mutex
}

func NewManager(filePath string) *Manager {
	return &Manager{
		filePath: filePath,
	}
}

type DNSConfig struct {
	Nameservers []string `json:"nameservers"`
	Search      []string `json:"search,omitempty"`
	Options     []string `json:"options,omitempty"`
}

func (m *Manager) GetAll() (*DNSConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.parseConfigFile()
}

func (m *Manager) Add(nameserver string) error {
	if !validator.IsValidIPAddress(nameserver) {
		return ErrInvalidIP
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	return m.modify(true, nameserver)
}

func (m *Manager) Remove(nameserver string) error {
	if !validator.IsValidIPAddress(nameserver) {
		return ErrInvalidIP
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	return m.modify(false, nameserver)
}

func (m *Manager) modify(add bool, target string) error {
	f, err := os.OpenFile(m.filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open resolv.conf: %w", err)
	}
	defer f.Close()

	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("failed to lock resolv.conf: %w", err)
	}
	defer unix.Flock(int(f.Fd()), unix.LOCK_UN)

	cfg, err := m.parseConfigFileWithReader(f)
	if err != nil {
		return err
	}

	if add {
		for _, ns := range cfg.Nameservers {
			if ns == target {
				return ErrAlreadyExists
			}
		}
		cfg.Nameservers = append(cfg.Nameservers, target)
	} else {
		found := false
		newNS := make([]string, 0, len(cfg.Nameservers))
		for _, ns := range cfg.Nameservers {
			if ns != target {
				newNS = append(newNS, ns)
			} else {
				found = true
			}
		}
		if !found {
			return ErrNotFound
		}
		cfg.Nameservers = newNS
	}

	return m.atomicWrite(cfg)
}

func (m *Manager) parseConfigFile() (*DNSConfig, error) {
	f, err := os.Open(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open resolv.conf: %w", err)
	}
	defer f.Close()
	return m.parseConfigFileWithReader(f)
}

func (m *Manager) parseConfigFileWithReader(f *os.File) (*DNSConfig, error) {
	f.Seek(0, 0)

	var nameservers, search, options []string
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "nameserver":
			if validator.IsValidIPAddress(parts[1]) {
				nameservers = append(nameservers, parts[1])
			}
		case "search":
			search = append(search, parts[1:]...)
		case "options":
			options = append(options, parts[1:]...)
		}
	}

	return &DNSConfig{
		Nameservers: nameservers,
		Search:      search,
		Options:     options,
	}, scanner.Err()
}

func (m *Manager) atomicWrite(cfg *DNSConfig) error {
	dir := filepath.Dir(m.filePath)
	tmp, err := os.CreateTemp(dir, "resolv.conf.tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	w := bufio.NewWriter(tmp)

	for _, ns := range cfg.Nameservers {
		fmt.Fprintf(w, "nameserver %s\n", ns)
	}
	if len(cfg.Search) > 0 {
		fmt.Fprintf(w, "search %s\n", strings.Join(cfg.Search, " "))
	}
	if len(cfg.Options) > 0 {
		fmt.Fprintf(w, "options %s\n", strings.Join(cfg.Options, " "))
	}

	if err := w.Flush(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	if err := os.Rename(tmpPath, m.filePath); err != nil {
		return fmt.Errorf("failed to replace resolv.conf: %w", err)
	}

	return nil
}
