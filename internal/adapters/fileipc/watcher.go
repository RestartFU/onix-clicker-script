package fileipc

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/restartfu/onix-winapi/internal/core"
)

type State struct {
	Enabled bool `json:"enabled"`
	CPS     int  `json:"cps"`
}

type Watcher struct {
	path     string
	interval time.Duration
}

func NewWatcher(path string, interval time.Duration) *Watcher {
	return &Watcher{path: path, interval: interval}
}

func (w *Watcher) Run(ctx context.Context, clicker *core.Clicker) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	var lastModTime time.Time
	var lastSize int64

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		info, err := os.Stat(w.path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		if info.ModTime().Equal(lastModTime) && info.Size() == lastSize {
			continue
		}

		payload, err := os.ReadFile(w.path)
		if err != nil {
			continue
		}

		payload = normalizePayload(payload)
		if len(payload) == 0 {
			continue
		}

		var state State
		if err := json.Unmarshal(payload, &state); err != nil {
			continue
		}

		if state.CPS > 0 {
			clicker.SetCPS(state.CPS)
		}
		clicker.SetEnabled(state.Enabled)

		lastModTime = info.ModTime()
		lastSize = info.Size()
	}
}

func normalizePayload(payload []byte) []byte {
	payload = bytes.TrimSpace(payload)
	if len(payload) == 0 {
		return payload
	}

	if len(payload) >= 3 && payload[0] == 0xEF && payload[1] == 0xBB && payload[2] == 0xBF {
		payload = bytes.TrimSpace(payload[3:])
	}

	if len(payload) >= 3 && payload[0] != '{' {
		length := int(payload[0]) | int(payload[1])<<8
		if length > 0 && 2+length <= len(payload) {
			candidate := bytes.TrimSpace(payload[2 : 2+length])
			if len(candidate) > 0 && candidate[0] == '{' {
				return candidate
			}
		}
	}

	return bytes.TrimSpace(payload)
}

func DefaultStatePath() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		home, _ := os.UserHomeDir()
		localAppData = filepath.Join(home, "AppData", "Local")
	}

	return filepath.Join(
		localAppData,
		"Packages",
		"Microsoft.MinecraftUWP_8wekyb3d8bbwe",
		"RoamingState",
		"OnixClient",
		"Scripts",
		"Data",
		"clicker",
		"state.json",
	)
}
