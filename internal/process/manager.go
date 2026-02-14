package process

import "strings"

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
	for _, p := range procs {
		if err := m.adapter.KillProcess(p.PID); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) KillAll(names []string) map[string]error {
	results := make(map[string]error)
	for _, name := range names {
		results[name] = m.KillByName(name)
	}
	return results
}
