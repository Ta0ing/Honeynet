package client

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func (c *Client) senseConfigPath() string {
	return filepath.Join(c.cfg.StateDir, "sense-config.json")
}

func (c *Client) loadSenseConfig() (protocol.SenseConfig, bool, error) {
	data, err := os.ReadFile(c.senseConfigPath())
	if errors.Is(err, os.ErrNotExist) {
		return protocol.SenseConfig{}, false, nil
	}
	if err != nil {
		return protocol.SenseConfig{}, false, err
	}
	var config protocol.SenseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return protocol.SenseConfig{}, false, err
	}
	return config, true, nil
}

func (c *Client) saveSenseConfig(config protocol.SenseConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(c.cfg.StateDir, 0700); err != nil {
		return err
	}
	path := c.senseConfigPath()
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0600); err != nil {
		return err
	}
	if err := os.Chmod(temporary, 0600); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
