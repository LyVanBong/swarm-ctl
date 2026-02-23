package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config là cấu hình chính của swarm-ctl tool
type Config struct {
	Version  string    `yaml:"version"`
	Clusters []Cluster `yaml:"clusters"`
	Current  string    `yaml:"current_cluster"`
}

// Cluster đại diện cho 1 Docker Swarm cluster
type Cluster struct {
	Name      string `yaml:"name"`
	MasterIP  string `yaml:"master_ip"`
	SSHKey    string `yaml:"ssh_key"`
	SSHUser   string `yaml:"ssh_user"`
	Domain    string `yaml:"domain"`
	DataRoot  string `yaml:"data_root"`
	CreatedAt string `yaml:"created_at"`
}

var defaultConfigPath = filepath.Join(os.Getenv("HOME"), ".swarm-ctl", "config.yml")

// Load đọc config từ file
func Load() (*Config, error) {
	return LoadFrom(defaultConfigPath)
}

// LoadFrom đọc config từ path cụ thể
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Tạo config mặc định nếu chưa có
			return &Config{Version: "1"}, nil
		}
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config format: %w", err)
	}
	return &cfg, nil
}

// Save lưu config xuống file
func (c *Config) Save() error {
	return c.SaveTo(defaultConfigPath)
}

// SaveTo lưu config xuống path cụ thể
func (c *Config) SaveTo(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// GetCurrentCluster trả về cluster đang active
func (c *Config) GetCurrentCluster() (*Cluster, error) {
	if c.Current == "" && len(c.Clusters) > 0 {
		return &c.Clusters[0], nil
	}
	for _, cl := range c.Clusters {
		if cl.Name == c.Current {
			return &cl, nil
		}
	}
	return nil, fmt.Errorf("no active cluster found — run: swarm-ctl cluster init")
}

// AddCluster thêm cluster mới vào config
func (c *Config) AddCluster(cl Cluster) {
	// Remove nếu đã tồn tại
	var filtered []Cluster
	for _, existing := range c.Clusters {
		if existing.Name != cl.Name {
			filtered = append(filtered, existing)
		}
	}
	c.Clusters = append(filtered, cl)
	if c.Current == "" {
		c.Current = cl.Name
	}
}
