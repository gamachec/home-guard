package process

import (
	"sync"
	"testing"
)

type mockAdapter struct {
	mu           sync.Mutex
	processes    []ProcessInfo
	applications []ProcessInfo
	killed       []uint32
}

func (m *mockAdapter) ListProcesses() ([]ProcessInfo, error) {
	return m.processes, nil
}

func (m *mockAdapter) ListApplications() ([]ProcessInfo, error) {
	return m.applications, nil
}

func (m *mockAdapter) KillProcess(pid uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.killed = append(m.killed, pid)
	return nil
}

func TestRunningApps(t *testing.T) {
	adapter := &mockAdapter{
		applications: []ProcessInfo{
			{PID: 5, Name: "notepad.exe", Path: `C:\Windows\notepad.exe`, Description: "Notepad"},
			{PID: 2, Name: "chrome.exe", Path: `C:\Program Files\Google\chrome.exe`, Description: "Google Chrome"},
		},
	}
	manager := NewManager(adapter)

	apps, err := manager.RunningApps()
	if err != nil {
		t.Fatalf("RunningApps() error = %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}
	if apps[0].PID != 2 || apps[1].PID != 5 {
		t.Errorf("apps not sorted by PID: got [%d, %d]", apps[0].PID, apps[1].PID)
	}
}

func TestFindByName(t *testing.T) {
	adapter := &mockAdapter{
		processes: []ProcessInfo{
			{PID: 1, Name: "roblox.exe"},
			{PID: 2, Name: "chrome.exe"},
			{PID: 3, Name: "notepad.exe"},
		},
	}
	manager := NewManager(adapter)

	results, err := manager.FindByName("roblox.exe")
	if err != nil {
		t.Fatalf("FindByName() error = %v", err)
	}
	if len(results) != 1 || results[0].PID != 1 {
		t.Fatalf("expected [roblox.exe PID=1], got %v", results)
	}
}

func TestFindByNameCaseInsensitive(t *testing.T) {
	adapter := &mockAdapter{
		processes: []ProcessInfo{
			{PID: 1, Name: "Roblox.exe"},
		},
	}
	manager := NewManager(adapter)

	results, err := manager.FindByName("roblox.exe")
	if err != nil {
		t.Fatalf("FindByName() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestKillByName(t *testing.T) {
	adapter := &mockAdapter{
		processes: []ProcessInfo{
			{PID: 100, Name: "roblox.exe"},
			{PID: 101, Name: "roblox.exe"},
		},
	}
	manager := NewManager(adapter)

	if err := manager.KillByName("roblox.exe"); err != nil {
		t.Fatalf("KillByName() error = %v", err)
	}

	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	if len(adapter.killed) != 2 {
		t.Fatalf("expected 2 kills, got %d", len(adapter.killed))
	}
}

func TestRunningFromBlacklist(t *testing.T) {
	adapter := &mockAdapter{
		processes: []ProcessInfo{
			{PID: 1, Name: "roblox.exe"},
			{PID: 2, Name: "roblox.exe"},
			{PID: 3, Name: "chrome.exe"},
			{PID: 4, Name: "notepad.exe"},
		},
	}
	manager := NewManager(adapter)

	running, err := manager.RunningFromBlacklist([]string{"roblox.exe", "chrome.exe", "fortnite.exe"})
	if err != nil {
		t.Fatalf("RunningFromBlacklist() error = %v", err)
	}
	if len(running) != 2 {
		t.Fatalf("expected 2 running apps, got %d: %v", len(running), running)
	}
	if running[0] != "roblox.exe" || running[1] != "chrome.exe" {
		t.Errorf("running = %v, want [roblox.exe chrome.exe]", running)
	}
}

func TestKillAll(t *testing.T) {
	adapter := &mockAdapter{
		processes: []ProcessInfo{
			{PID: 1, Name: "roblox.exe"},
			{PID: 2, Name: "chrome.exe"},
		},
	}
	manager := NewManager(adapter)

	results := manager.KillAll([]string{"roblox.exe", "chrome.exe"})
	for name, err := range results {
		if err != nil {
			t.Errorf("KillAll[%s] unexpected error: %v", name, err)
		}
	}

	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	if len(adapter.killed) != 2 {
		t.Fatalf("expected 2 kills, got %d", len(adapter.killed))
	}
}
