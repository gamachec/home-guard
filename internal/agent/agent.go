package agent

import (
	"context"
	"sync"
	"time"

	"home-guard/internal/config"
	"home-guard/internal/process"
)

type Mode string

const (
	ModeActive  Mode = "ACTIVE"
	ModeBlocked Mode = "BLOCKED"
)

type Agent struct {
	mu         sync.RWMutex
	mode       Mode
	blacklist  []string
	cfg        *config.Config
	configPath string
	manager    *process.Manager
	onPublish  func(mode Mode)
	stopBlock  context.CancelFunc
	killDelay  func() time.Duration
}

func New(
	manager *process.Manager,
	cfg *config.Config,
	configPath string,
	onPublish func(mode Mode),
) *Agent {
	return &Agent{
		mode:       ModeActive,
		blacklist:  cfg.Blacklist,
		cfg:        cfg,
		configPath: configPath,
		manager:    manager,
		onPublish:  onPublish,
		killDelay:  defaultKillDelay,
	}
}

func defaultKillDelay() time.Duration {
	return time.Second
}

func (a *Agent) SetMode(ctx context.Context, mode Mode) {
	a.mu.Lock()

	previous := a.mode
	a.mode = mode

	if previous == ModeBlocked && mode != ModeBlocked && a.stopBlock != nil {
		a.stopBlock()
		a.stopBlock = nil
	}

	var blockCtx context.Context
	if mode == ModeBlocked && previous != ModeBlocked {
		blockCtx, a.stopBlock = context.WithCancel(ctx)
	}

	a.mu.Unlock()

	if blockCtx != nil {
		go a.runKillLoop(blockCtx)
	}

	if a.onPublish != nil {
		a.onPublish(mode)
	}
}

func (a *Agent) SetBlacklist(apps []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.blacklist = apps
	a.cfg.Blacklist = apps

	return config.Save(a.configPath, a.cfg)
}

func (a *Agent) Blacklist() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]string, len(a.blacklist))
	copy(result, a.blacklist)
	return result
}

func (a *Agent) runKillLoop(ctx context.Context) {
	for {
		a.mu.RLock()
		blacklist := make([]string, len(a.blacklist))
		copy(blacklist, a.blacklist)
		a.mu.RUnlock()

		a.manager.KillAll(blacklist)

		select {
		case <-ctx.Done():
			return
		case <-time.After(a.killDelay()):
		}
	}
}
