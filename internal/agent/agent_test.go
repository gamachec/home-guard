package agent

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"home-guard/internal/config"
	"home-guard/internal/process"
)

type mockAdapter struct {
	mu     sync.Mutex
	procs  []process.ProcessInfo
	killed []string
}

func (m *mockAdapter) ListProcesses() ([]process.ProcessInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.procs, nil
}

func (m *mockAdapter) ListApplications() ([]process.ProcessInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.procs, nil
}

func (m *mockAdapter) KillProcess(pid uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.procs {
		if p.PID == pid {
			m.killed = append(m.killed, p.Name)
		}
	}
	return nil
}

func newTestAgent(cfg *config.Config, configPath string, adapter *mockAdapter, onPublish func(Mode)) *Agent {
	manager := process.NewManager(adapter)
	a := New(manager, cfg, configPath, onPublish)
	a.killDelay = func() time.Duration { return 10 * time.Millisecond }
	a.scanDelay = func() time.Duration { return 10 * time.Millisecond }
	return a
}

func TestSetModeBlockedKillsBlacklist(t *testing.T) {
	adapter := &mockAdapter{
		procs: []process.ProcessInfo{
			{PID: 1, Name: "game.exe"},
		},
	}
	cfg := &config.Config{Blacklist: []string{"game.exe"}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := newTestAgent(cfg, "", adapter, nil)
	a.SetMode(ctx, ModeBlocked)

	time.Sleep(100 * time.Millisecond)
	cancel()

	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	if len(adapter.killed) == 0 {
		t.Error("expected at least one process to be killed")
	}
	if adapter.killed[0] != "game.exe" {
		t.Errorf("killed[0] = %q, want %q", adapter.killed[0], "game.exe")
	}
}

func TestSetModeBlockedStopsOnModeChange(t *testing.T) {
	adapter := &mockAdapter{
		procs: []process.ProcessInfo{
			{PID: 1, Name: "game.exe"},
		},
	}
	cfg := &config.Config{Blacklist: []string{"game.exe"}}

	ctx := context.Background()
	a := newTestAgent(cfg, "", adapter, nil)
	a.SetMode(ctx, ModeBlocked)

	time.Sleep(50 * time.Millisecond)

	a.SetMode(ctx, ModeActive)

	adapter.mu.Lock()
	killsAfterSwitch := len(adapter.killed)
	adapter.mu.Unlock()

	time.Sleep(100 * time.Millisecond)

	adapter.mu.Lock()
	killsAfterWait := len(adapter.killed)
	adapter.mu.Unlock()

	if killsAfterWait > killsAfterSwitch+1 {
		t.Errorf("kill loop kept running after mode change: %d kills vs %d", killsAfterWait, killsAfterSwitch)
	}
}

func TestSetModePublishesState(t *testing.T) {
	adapter := &mockAdapter{}
	cfg := &config.Config{}

	var published Mode
	a := newTestAgent(cfg, "", adapter, func(m Mode) {
		published = m
	})

	a.SetMode(context.Background(), ModeBlocked)

	if published != ModeBlocked {
		t.Errorf("published = %q, want %q", published, ModeBlocked)
	}

	a.SetMode(context.Background(), ModeActive)

	if published != ModeActive {
		t.Errorf("published = %q, want %q", published, ModeActive)
	}
}

func TestStartPublishesRunningApps(t *testing.T) {
	adapter := &mockAdapter{
		procs: []process.ProcessInfo{
			{PID: 1, Name: "game.exe", Path: `C:\game.exe`},
			{PID: 2, Name: "chrome.exe", Path: `C:\chrome.exe`, Description: "Google Chrome"},
		},
	}
	cfg := &config.Config{}

	ch := make(chan []process.ProcessInfo, 1)
	a := newTestAgent(cfg, "", adapter, nil)
	a.SetOnPublishRunning(func(apps []process.ProcessInfo) {
		select {
		case ch <- apps:
		default:
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.Start(ctx)

	select {
	case apps := <-ch:
		if len(apps) != 2 {
			t.Fatalf("expected 2 apps, got %d: %v", len(apps), apps)
		}
		if apps[0].Name != "game.exe" || apps[1].Name != "chrome.exe" {
			t.Errorf("published apps = %v", apps)
		}
		if apps[1].Description != "Google Chrome" {
			t.Errorf("apps[1].Description = %q, want %q", apps[1].Description, "Google Chrome")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout: aucune app publiée")
	}
}

func TestSetBlacklist(t *testing.T) {
	f, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"broker":"localhost","port":1883,"client_id":"pc","blacklist":[]}`)
	f.Close()
	defer os.Remove(f.Name())

	adapter := &mockAdapter{}
	cfg := &config.Config{Blacklist: []string{}}

	a := newTestAgent(cfg, f.Name(), adapter, nil)

	apps := []string{"game.exe", "browser.exe"}
	if err := a.SetBlacklist(apps); err != nil {
		t.Fatalf("SetBlacklist() error = %v", err)
	}

	bl := a.Blacklist()
	if len(bl) != 2 {
		t.Fatalf("Blacklist length = %d, want 2", len(bl))
	}
	if bl[0] != "game.exe" || bl[1] != "browser.exe" {
		t.Errorf("Blacklist = %v, want [game.exe browser.exe]", bl)
	}

	loaded, err := config.Load(f.Name())
	if err != nil {
		t.Fatalf("config.Load() after SetBlacklist: %v", err)
	}
	if len(loaded.Blacklist) != 2 || loaded.Blacklist[0] != "game.exe" {
		t.Errorf("persisted blacklist = %v, want [game.exe browser.exe]", loaded.Blacklist)
	}
}

func TestSetBlacklistUpdatesKillLoop(t *testing.T) {
	adapter := &mockAdapter{
		procs: []process.ProcessInfo{
			{PID: 1, Name: "newgame.exe"},
		},
	}
	cfg := &config.Config{Blacklist: []string{}}

	f, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"broker":"localhost","port":1883,"client_id":"pc","blacklist":[]}`)
	f.Close()
	defer os.Remove(f.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := newTestAgent(cfg, f.Name(), adapter, nil)
	a.SetMode(ctx, ModeBlocked)

	time.Sleep(30 * time.Millisecond)

	adapter.mu.Lock()
	killsBefore := len(adapter.killed)
	adapter.mu.Unlock()

	_ = a.SetBlacklist([]string{"newgame.exe"})

	time.Sleep(100 * time.Millisecond)
	cancel()

	adapter.mu.Lock()
	killsAfter := len(adapter.killed)
	adapter.mu.Unlock()

	if killsAfter <= killsBefore {
		t.Error("expected new blacklist to trigger kills")
	}
}
