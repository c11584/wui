package xray

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type Process struct {
	ID      string
	Port    int
	cmd     *exec.Cmd
	process *os.Process
	mu      sync.RWMutex
	running bool
}

type Manager struct {
	binPath    string
	configPath string
	processes  map[string]*Process
	mu         sync.RWMutex
}

func NewManager(binPath, configPath string) *Manager {
	return &Manager{
		binPath:    binPath,
		configPath: configPath,
		processes:  make(map[string]*Process),
	}
}

// Start starts Xray process with given config
func (m *Manager) Start(id string, config string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	if p, exists := m.processes[id]; exists && p.running {
		return fmt.Errorf("process %s is already running", id)
	}

	// Write config file
	configFile := filepath.Join(m.configPath, fmt.Sprintf("%s.json", id))
	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	// Create command
	cmd := exec.Command(m.binPath, "run", "-c", configFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start xray: %v", err)
	}

	// Store process
	m.processes[id] = &Process{
		ID:      id,
		cmd:     cmd,
		process: cmd.Process,
		running: true,
	}

	// Wait for process in goroutine
	go func() {
		cmd.Wait()
		m.mu.Lock()
		if p, exists := m.processes[id]; exists {
			p.running = false
		}
		m.mu.Unlock()
	}()

	return nil
}

// Stop stops Xray process
func (m *Manager) Stop(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.processes[id]
	if !exists {
		return nil
	}

	if p.running {
		// Kill process
		if err := p.process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %v", err)
		}
		p.running = false
	}

	delete(m.processes, id)

	configFile := filepath.Join(m.configPath, fmt.Sprintf("%s.json", id))
	os.Remove(configFile)

	return nil
}

// Restart restarts Xray process
func (m *Manager) Restart(id string, config string) error {
	// Stop if running
	m.Stop(id)

	// Wait a bit
	time.Sleep(500 * time.Millisecond)

	// Start again
	return m.Start(id, config)
}

// IsRunning checks if process is running
func (m *Manager) IsRunning(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, exists := m.processes[id]
	return exists && p.running
}

// GetRunningProcesses returns all running process IDs
func (m *Manager) GetRunningProcesses() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []string
	for id, p := range m.processes {
		if p.running {
			result = append(result, id)
		}
	}
	return result
}

// StopAll stops all running processes
func (m *Manager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, p := range m.processes {
		if p.running {
			p.process.Kill()
			p.running = false
		}
		delete(m.processes, id)
	}

	return nil
}

// GracefulShutdown gracefully shuts down all processes
func (m *Manager) GracefulShutdown(ctx context.Context) error {
	done := make(chan error, 1)

	go func() {
		done <- m.StopAll()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
