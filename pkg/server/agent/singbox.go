package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	configv1 "sing-box-web/pkg/config/v1"
	pbv1 "sing-box-web/pkg/pb/v1"
)

// SingboxManager manages the sing-box process
type SingboxManager struct {
	config configv1.AgentConfig
	logger *zap.Logger

	// Process management
	cmd       *exec.Cmd
	pid       int
	processMu sync.RWMutex

	// Configuration
	configPath string
	configMu   sync.RWMutex

	// Traffic data
	trafficData map[string]*pbv1.UserTraffic
	trafficMu   sync.RWMutex

	// Shutdown
	shutdownCtx context.Context
	shutdown    context.CancelFunc
}

// SingboxConfig represents the sing-box configuration
type SingboxConfig struct {
	Log struct {
		Level     string `json:"level"`
		Timestamp bool   `json:"timestamp"`
	} `json:"log"`
	Inbounds []struct {
		Type   string `json:"type"`
		Tag    string `json:"tag"`
		Listen string `json:"listen"`
		Port   int    `json:"port"`
		Users  []struct {
			UUID     string `json:"uuid"`
			Username string `json:"username"`
		} `json:"users,omitempty"`
	} `json:"inbounds"`
	Outbounds []struct {
		Type string `json:"type"`
		Tag  string `json:"tag"`
	} `json:"outbounds"`
}

// NewSingboxManager creates a new sing-box manager
func NewSingboxManager(config configv1.AgentConfig, logger *zap.Logger) *SingboxManager {
	shutdownCtx, shutdown := context.WithCancel(context.Background())

	return &SingboxManager{
		config:      config,
		logger:      logger.Named("singbox"),
		configPath:  filepath.Join(config.SingBox.WorkingDir, "config.json"),
		trafficData: make(map[string]*pbv1.UserTraffic),
		shutdownCtx: shutdownCtx,
		shutdown:    shutdown,
	}
}

// Start starts the sing-box manager
func (s *SingboxManager) Start(ctx context.Context) error {
	s.logger.Info("starting sing-box manager")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(s.config.SingBox.WorkingDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Initialize configuration
	if err := s.initializeConfig(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Start sing-box process
	if err := s.startSingboxProcess(); err != nil {
		return fmt.Errorf("failed to start sing-box process: %w", err)
	}

	// Start monitoring
	go s.monitorProcess()
	go s.trafficCollectionLoop()

	return nil
}

// Stop stops the sing-box manager
func (s *SingboxManager) Stop(ctx context.Context) error {
	s.logger.Info("stopping sing-box manager")

	// Cancel background tasks
	s.shutdown()

	// Stop sing-box process
	if err := s.stopSingboxProcess(); err != nil {
		s.logger.Error("failed to stop sing-box process", zap.Error(err))
	}

	return nil
}

// initializeConfig initializes the sing-box configuration
func (s *SingboxManager) initializeConfig() error {
	s.logger.Info("initializing sing-box configuration")

	// Create default configuration
	config := SingboxConfig{
		Log: struct {
			Level     string `json:"level"`
			Timestamp bool   `json:"timestamp"`
		}{
			Level:     "info",
			Timestamp: true,
		},
		Inbounds: []struct {
			Type   string `json:"type"`
			Tag    string `json:"tag"`
			Listen string `json:"listen"`
			Port   int    `json:"port"`
			Users  []struct {
				UUID     string `json:"uuid"`
				Username string `json:"username"`
			} `json:"users,omitempty"`
		}{
			{
				Type:   "vless",
				Tag:    "vless-in",
				Listen: "0.0.0.0",
				Port:   s.config.SingBox.ClashAPI.Port,
				Users:  []struct {
					UUID     string `json:"uuid"`
					Username string `json:"username"`
				}{},
			},
		},
		Outbounds: []struct {
			Type string `json:"type"`
			Tag  string `json:"tag"`
		}{
			{
				Type: "direct",
				Tag:  "direct",
			},
		},
	}

	// Write configuration to file
	return s.writeConfig(config)
}

// writeConfig writes the configuration to file
func (s *SingboxManager) writeConfig(config SingboxConfig) error {
	s.configMu.Lock()
	defer s.configMu.Unlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := ioutil.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	s.logger.Debug("configuration written", zap.String("path", s.configPath))
	return nil
}

// readConfig reads the configuration from file
func (s *SingboxManager) readConfig() (*SingboxConfig, error) {
	s.configMu.RLock()
	defer s.configMu.RUnlock()

	data, err := ioutil.ReadFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config SingboxConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// startSingboxProcess starts the sing-box process
func (s *SingboxManager) startSingboxProcess() error {
	s.processMu.Lock()
	defer s.processMu.Unlock()

	if s.cmd != nil {
		return fmt.Errorf("sing-box process is already running")
	}

	s.logger.Info("starting sing-box process", zap.String("config", s.configPath))

	// Create command
	s.cmd = exec.Command("sing-box", "run", "-c", s.configPath)
	s.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Start process
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start sing-box process: %w", err)
	}

	s.pid = s.cmd.Process.Pid
	s.logger.Info("sing-box process started", zap.Int("pid", s.pid))

	return nil
}

// stopSingboxProcess stops the sing-box process
func (s *SingboxManager) stopSingboxProcess() error {
	s.processMu.Lock()
	defer s.processMu.Unlock()

	if s.cmd == nil {
		return nil
	}

	s.logger.Info("stopping sing-box process", zap.Int("pid", s.pid))

	// Send SIGTERM
	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		s.logger.Error("failed to send SIGTERM", zap.Error(err))
		// Force kill
		s.cmd.Process.Kill()
	}

	// Wait for process to exit
	s.cmd.Wait()

	s.cmd = nil
	s.pid = 0

	s.logger.Info("sing-box process stopped")
	return nil
}

// restartSingboxProcess restarts the sing-box process
func (s *SingboxManager) restartSingboxProcess() error {
	s.logger.Info("restarting sing-box process")

	// Stop current process
	if err := s.stopSingboxProcess(); err != nil {
		s.logger.Error("failed to stop sing-box process", zap.Error(err))
	}

	// Start new process
	if err := s.startSingboxProcess(); err != nil {
		return fmt.Errorf("failed to start sing-box process: %w", err)
	}

	return nil
}

// monitorProcess monitors the sing-box process
func (s *SingboxManager) monitorProcess() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdownCtx.Done():
			return
		case <-ticker.C:
			s.checkProcessHealth()
		}
	}
}

// checkProcessHealth checks if the sing-box process is healthy
func (s *SingboxManager) checkProcessHealth() {
	s.processMu.RLock()
	cmd := s.cmd
	s.processMu.RUnlock()

	if cmd == nil {
		s.logger.Warn("sing-box process is not running, attempting to restart")
		if err := s.startSingboxProcess(); err != nil {
			s.logger.Error("failed to restart sing-box process", zap.Error(err))
		}
		return
	}

	// Check if process is still alive
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		s.logger.Warn("sing-box process has exited, attempting to restart")
		if err := s.restartSingboxProcess(); err != nil {
			s.logger.Error("failed to restart sing-box process", zap.Error(err))
		}
	}
}

// trafficCollectionLoop collects traffic data
func (s *SingboxManager) trafficCollectionLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdownCtx.Done():
			return
		case <-ticker.C:
			s.collectTrafficData()
		}
	}
}

// collectTrafficData collects traffic data from sing-box
func (s *SingboxManager) collectTrafficData() {
	// This is a placeholder implementation
	// In a real implementation, you would query sing-box API or parse logs
	s.trafficMu.Lock()
	defer s.trafficMu.Unlock()

	// Generate some mock traffic data
	s.trafficData["user1"] = &pbv1.UserTraffic{
		UserId:      "1",
		UploadBytes:   1024 * 1024,     // 1MB
		DownloadBytes: 1024 * 1024 * 5, // 5MB
	}
	s.trafficData["user2"] = &pbv1.UserTraffic{
		UserId:      "2",
		UploadBytes:   1024 * 1024 * 2, // 2MB
		DownloadBytes: 1024 * 1024 * 3, // 3MB
	}
}

// GetPID returns the sing-box process PID
func (s *SingboxManager) GetPID() int {
	s.processMu.RLock()
	defer s.processMu.RUnlock()
	return s.pid
}

// GetTrafficData returns and clears the traffic data
func (s *SingboxManager) GetTrafficData() []*pbv1.UserTraffic {
	s.trafficMu.Lock()
	defer s.trafficMu.Unlock()

	if len(s.trafficData) == 0 {
		return nil
	}

	// Convert map to slice
	data := make([]*pbv1.UserTraffic, 0, len(s.trafficData))
	for _, traffic := range s.trafficData {
		data = append(data, traffic)
	}

	// Clear the data
	s.trafficData = make(map[string]*pbv1.UserTraffic)

	return data
}

// AddUser adds a user to the sing-box configuration
func (s *SingboxManager) AddUser(userID string, parameters map[string]string) error {
	s.logger.Info("adding user to sing-box", zap.String("user_id", userID))

	// Read current configuration
	config, err := s.readConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Add user to inbound configuration
	uuid := parameters["uuid"]
	if uuid == "" {
		uuid = "user-" + userID + "-uuid" // Generate UUID
	}

	for i := range config.Inbounds {
		if config.Inbounds[i].Type == "vless" {
			config.Inbounds[i].Users = append(config.Inbounds[i].Users, struct {
				UUID     string `json:"uuid"`
				Username string `json:"username"`
			}{
				UUID:     uuid,
				Username: "user" + userID,
			})
			break
		}
	}

	// Write updated configuration
	if err := s.writeConfig(*config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Restart sing-box to apply changes
	return s.restartSingboxProcess()
}

// RemoveUser removes a user from the sing-box configuration
func (s *SingboxManager) RemoveUser(userID string) error {
	s.logger.Info("removing user from sing-box", zap.String("user_id", userID))

	// Read current configuration
	config, err := s.readConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Remove user from inbound configuration
	username := "user" + userID
	for i := range config.Inbounds {
		if config.Inbounds[i].Type == "vless" {
			newUsers := make([]struct {
				UUID     string `json:"uuid"`
				Username string `json:"username"`
			}, 0)
			for _, user := range config.Inbounds[i].Users {
				if user.Username != username {
					newUsers = append(newUsers, user)
				}
			}
			config.Inbounds[i].Users = newUsers
			break
		}
	}

	// Write updated configuration
	if err := s.writeConfig(*config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Restart sing-box to apply changes
	return s.restartSingboxProcess()
}

// UpdateUser updates a user in the sing-box configuration
func (s *SingboxManager) UpdateUser(userID string, parameters map[string]string) error {
	s.logger.Info("updating user in sing-box", zap.String("user_id", userID))

	// For now, just restart the process
	// In a real implementation, you would update the user configuration
	return s.restartSingboxProcess()
}

// ResetTraffic resets traffic for a user
func (s *SingboxManager) ResetTraffic(userID string) error {
	s.logger.Info("resetting traffic for user", zap.String("user_id", userID))

	// Clear traffic data for the user
	s.trafficMu.Lock()
	defer s.trafficMu.Unlock()

	delete(s.trafficData, "user"+userID)

	return nil
}