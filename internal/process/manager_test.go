package process

import (
	"sync"
	"testing"
)

type mockAdapter struct {
	mu        sync.Mutex
	processes []ProcessInfo
	killed    []uint32
}

func (m *mockAdapter) ListProcesses() ([]ProcessInfo, error) {
	return m.processes, nil
}

func (m *mockAdapter) KillProcess(pid uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.killed = append(m.killed, pid)
	return nil
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
