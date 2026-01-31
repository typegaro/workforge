package plugin

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PluginInfo struct {
	Name       string
	SocketPath string
	Process    *os.Process
}

type PluginService struct {
	pluginsDir string
	socketsDir string
	plugins    map[string]*PluginInfo
	mu         sync.RWMutex
	requestID  int
}

func NewPluginService(pluginsDir, socketsDir string) *PluginService {
	return &PluginService{
		pluginsDir: pluginsDir,
		socketsDir: socketsDir,
		plugins:    make(map[string]*PluginInfo),
	}
}

func DefaultPluginsDir() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "workforge", "plugins")
}

func DefaultSocketsDir() string {
	return filepath.Join(os.TempDir(), "workforge-plugins")
}

func (s *PluginService) Wakeup(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	socketPath := filepath.Join(s.socketsDir, name+".sock")

	if info, exists := s.plugins[name]; exists {
		if s.isAlive(info) {
			return nil
		}
		delete(s.plugins, name)
	}

	if s.isSocketAlive(socketPath) {
		s.plugins[name] = &PluginInfo{
			Name:       name,
			SocketPath: socketPath,
			Process:    nil,
		}
		return nil
	}

	pluginDir := filepath.Join(s.pluginsDir, name)
	manifest, err := LoadManifest(pluginDir)
	if err != nil {
		return fmt.Errorf("load plugin manifest: %w", err)
	}

	entrypoint := filepath.Join(pluginDir, manifest.Entrypoint)
	if _, err := os.Stat(entrypoint); err != nil {
		return fmt.Errorf("plugin %q not found: %w", name, err)
	}

	if err := os.MkdirAll(s.socketsDir, 0o755); err != nil {
		return fmt.Errorf("create sockets dir: %w", err)
	}

	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cleanup old socket: %w", err)
	}

	cmd := exec.Command(manifest.Runtime, entrypoint, socketPath)
	cmd.Dir = pluginDir
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start plugin %q: %w", name, err)
	}

	if err := s.waitForSocket(socketPath, 5*time.Second); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("plugin %q failed to start: %w", name, err)
	}

	s.plugins[name] = &PluginInfo{
		Name:       name,
		SocketPath: socketPath,
		Process:    cmd.Process,
	}

	return nil
}

func (s *PluginService) Call(name, method string, params interface{}) (json.RawMessage, error) {
	s.mu.RLock()
	info, exists := s.plugins[name]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %q not running", name)
	}

	conn, err := net.DialTimeout("unix", info.SocketPath, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect to plugin %q: %w", name, err)
	}
	defer conn.Close()

	s.mu.Lock()
	s.requestID++
	reqID := s.requestID
	s.mu.Unlock()

	req := Request{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  method,
		Params:  params,
	}

	conn.SetDeadline(time.Now().Add(30 * time.Second))

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("plugin error [%d]: %s", resp.Error.Code, resp.Error.Message)
	}

	return resp.Result, nil
}

func (s *PluginService) Kill(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.plugins[name]
	if !exists {
		return nil
	}

	s.callShutdown(info)

	if info.Process != nil {
		info.Process.Kill()
		info.Process.Wait()
	}

	os.Remove(info.SocketPath)
	delete(s.plugins, name)

	return nil
}

func (s *PluginService) KillAll() {
	s.mu.Lock()
	names := make([]string, 0, len(s.plugins))
	for name := range s.plugins {
		names = append(names, name)
	}
	s.mu.Unlock()

	for _, name := range names {
		s.Kill(name)
	}
}

// WakeupResult contains the result of an async wakeup operation
type WakeupResult struct {
	Name  string
	Error error
}

// WakeupAsync starts a plugin asynchronously and returns immediately.
// The result is sent to the returned channel when the wakeup completes or fails.
func (s *PluginService) WakeupAsync(name string) <-chan WakeupResult {
	ch := make(chan WakeupResult, 1)
	go func() {
		err := s.Wakeup(name)
		ch <- WakeupResult{Name: name, Error: err}
		close(ch)
	}()
	return ch
}

func (s *PluginService) IsRunning(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.plugins[name]
	if !exists {
		return false
	}
	return s.isAlive(info)
}

func (s *PluginService) Ping(name string) (bool, error) {
	socketPath := filepath.Join(s.socketsDir, name+".sock")
	conn, err := net.DialTimeout("unix", socketPath, 500*time.Millisecond)
	if err != nil {
		return false, nil
	}
	defer conn.Close()

	req := Request{JSONRPC: "2.0", ID: 1, Method: "on_ping"}
	conn.SetDeadline(time.Now().Add(2 * time.Second))

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return false, nil
	}

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return false, nil
	}

	return resp.Error == nil, nil
}

func (s *PluginService) ListRunning() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.plugins))
	for name := range s.plugins {
		names = append(names, name)
	}
	return names
}

func (s *PluginService) RunHook(registry *PluginRegistryService, hook string, payload interface{}) map[string]HookResult {
	plugins, err := registry.List()
	if err != nil {
		return map[string]HookResult{"registry": {Error: err}}
	}

	results := make(map[string]HookResult)
	for _, p := range plugins {
		if !hasHook(p.Hooks, hook) {
			continue
		}

		if err := s.Wakeup(p.Name); err != nil {
			results[p.Name] = HookResult{Error: err}
			continue
		}

		resp, err := s.Call(p.Name, hook, payload)
		if err != nil {
			results[p.Name] = HookResult{Error: err}
			continue
		}

		results[p.Name] = HookResult{Response: resp}
	}

	return results
}

type HookResult struct {
	Response json.RawMessage
	Error    error
}

func hasHook(hooks []string, target string) bool {
	for _, h := range hooks {
		if h == target {
			return true
		}
	}
	return false
}

func (s *PluginService) isAlive(info *PluginInfo) bool {
	return s.isSocketAlive(info.SocketPath)
}

func (s *PluginService) isSocketAlive(socketPath string) bool {
	conn, err := net.DialTimeout("unix", socketPath, 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (s *PluginService) waitForSocket(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("unix", path, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for socket")
}

func (s *PluginService) callShutdown(info *PluginInfo) {
	conn, err := net.DialTimeout("unix", info.SocketPath, 1*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()

	req := Request{JSONRPC: "2.0", ID: 0, Method: "shutdown"}
	conn.SetDeadline(time.Now().Add(2 * time.Second))
	json.NewEncoder(conn).Encode(req)
}
