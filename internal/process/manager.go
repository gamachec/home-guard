package process

import (
	"strings"
	"sync"
)

type ProcessInfo struct {
	PID  uint32
	Name string
}

type OSAdapter interface {
	ListProcesses() ([]ProcessInfo, error)
	KillProcess(pid uint32) error
}

type Manager struct {
	adapter OSAdapter
}

func NewManager(adapter OSAdapter) *Manager {
	return &Manager{adapter: adapter}
}

func (m *Manager) FindByName(name string) ([]ProcessInfo, error) {
	all, err := m.adapter.ListProcesses()
	if err != nil {
		return nil, err
	}

	var matches []ProcessInfo
	for _, p := range all {
		if strings.EqualFold(p.Name, name) {
			matches = append(matches, p)
		}
	}
	return matches, nil
}

func (m *Manager) KillByName(name string) error {
	procs, err := m.FindByName(name)
	if err != nil {
		return err
	}

	errs := make(chan error, len(procs))
	var wg sync.WaitGroup

	for _, p := range procs {
		wg.Add(1)
		go func(pid uint32) {
			defer wg.Done()
			if err := m.adapter.KillProcess(pid); err != nil {
				errs <- err
			}
		}(p.PID)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		return err
	}
	return nil
}

func (m *Manager) RunningFromBlacklist(blacklist []string) ([]string, error) {
	all, err := m.adapter.ListProcesses()
	if err != nil {
		return nil, err
	}

	running := make([]string, 0)
	for _, name := range blacklist {
		for _, p := range all {
			if strings.EqualFold(p.Name, name) {
				running = append(running, name)
				break
			}
		}
	}

	return running, nil
}

func (m *Manager) KillAll(names []string) map[string]error {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results = make(map[string]error)
	)

	for _, name := range names {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			err := m.KillByName(n)
			mu.Lock()
			results[n] = err
			mu.Unlock()
		}(name)
	}

	wg.Wait()
	return results
}
