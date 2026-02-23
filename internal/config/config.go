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

// Subdomains chứa tên miền phụ tùy chỉnh cho các service hạ tầng
type Subdomains struct {
	Traefik   string `yaml:"traefik,omitempty"`
	Portainer string `yaml:"portainer,omitempty"`
}

// ACME chứa cấu hình Let's Encrypt SSL
type ACME struct {
	Email     string `yaml:"email,omitempty"`     // Email đăng ký SSL (mặc định: admin@{domain})
	Challenge string `yaml:"challenge,omitempty"` // tls | http (mặc định: tls)
}

// TraefikAuth chứa cấu hình BasicAuth cho Traefik Dashboard
type TraefikAuth struct {
	Username string `yaml:"username,omitempty"` // mặc định: admin
	Password string `yaml:"password,omitempty"` // BCrypt hash hoặc plain text (sẽ tự hash)
}

// Registry chứa cấu hình Docker Registry riêng tư
type Registry struct {
	Server   string `yaml:"server,omitempty"`   // mặc định: docker.io
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// BackupS3 chứa cấu hình S3 backup
type BackupS3 struct {
	Endpoint  string `yaml:"endpoint,omitempty"`   // S3 endpoint (vd: s3.amazonaws.com)
	Bucket    string `yaml:"bucket,omitempty"`      // Tên bucket
	AccessKey string `yaml:"access_key,omitempty"`
	SecretKey string `yaml:"secret_key,omitempty"`
	Region    string `yaml:"region,omitempty"`      // mặc định: us-east-1
	Schedule  string `yaml:"schedule,omitempty"`    // Cron schedule (vd: 0 2 * * *)
}

// AlertTelegram chứa cấu hình Telegram alert
type AlertTelegram struct {
	BotToken string `yaml:"bot_token,omitempty"` // Telegram Bot Token
	ChatID   string `yaml:"chat_id,omitempty"`   // Telegram Chat ID
	Enabled  bool   `yaml:"enabled,omitempty"`
}

// Cluster đại diện cho 1 Docker Swarm cluster
type Cluster struct {
	Name       string        `yaml:"name"`
	MasterIP   string        `yaml:"master_ip"`
	SSHKey     string        `yaml:"ssh_key"`
	SSHUser    string        `yaml:"ssh_user"`
	Domain     string        `yaml:"domain"`
	DataRoot   string        `yaml:"data_root"`
	CreatedAt  string        `yaml:"created_at"`
	Subdomains Subdomains    `yaml:"subdomains,omitempty"`
	ACME       ACME          `yaml:"acme,omitempty"`
	Auth       TraefikAuth   `yaml:"traefik_auth,omitempty"`
	Registry   Registry      `yaml:"registry,omitempty"`
	Backup     BackupS3      `yaml:"backup_s3,omitempty"`
	Alert      AlertTelegram `yaml:"alert_telegram,omitempty"`
}

// ── Helper: Lấy giá trị cấu hình (có fallback mặc định) ──

func (c *Cluster) GetTraefikHost() string {
	sub := c.Subdomains.Traefik
	if sub == "" {
		sub = "traefik"
	}
	return sub + "." + c.Domain
}

func (c *Cluster) GetPortainerHost() string {
	sub := c.Subdomains.Portainer
	if sub == "" {
		sub = "portainer"
	}
	return sub + "." + c.Domain
}

func (c *Cluster) GetACMEEmail() string {
	if c.ACME.Email != "" {
		return c.ACME.Email
	}
	return "admin@" + c.Domain
}

func (c *Cluster) GetACMEChallenge() string {
	if c.ACME.Challenge != "" {
		return c.ACME.Challenge
	}
	return "tls"
}

func (c *Cluster) GetAuthUsername() string {
	if c.Auth.Username != "" {
		return c.Auth.Username
	}
	return "admin"
}

func (c *Cluster) GetRegistryServer() string {
	if c.Registry.Server != "" {
		return c.Registry.Server
	}
	return "docker.io"
}

func (c *Cluster) GetBackupRegion() string {
	if c.Backup.Region != "" {
		return c.Backup.Region
	}
	return "us-east-1"
}

func (c *Cluster) GetBackupSchedule() string {
	if c.Backup.Schedule != "" {
		return c.Backup.Schedule
	}
	return "0 2 * * *"
}

// ── CRUD Operations ──

// UpdateCluster cập nhật cluster theo tên
func (c *Config) UpdateCluster(updated Cluster) {
	for i, cl := range c.Clusters {
		if cl.Name == updated.Name {
			c.Clusters[i] = updated
			return
		}
	}
}

// AllConfigKeys trả về danh sách tất cả key hợp lệ kèm mô tả
func AllConfigKeys() map[string]string {
	return map[string]string{
		"domain":                "Domain chính của cluster",
		"cluster-name":         "Tên hiển thị cluster",
		"data-root":            "Thư mục gốc lưu dữ liệu trên server",
		"ssh-user":             "SSH username",
		"ssh-key":              "Đường dẫn SSH private key",
		"traefik-subdomain":    "Subdomain cho Traefik Dashboard",
		"portainer-subdomain":  "Subdomain cho Portainer",
		"acme-email":           "Email đăng ký SSL Let's Encrypt",
		"acme-challenge":       "Kiểu xác thực SSL: tls | http",
		"traefik-auth-user":    "Username BasicAuth Traefik Dashboard",
		"traefik-auth-pass":    "Password BasicAuth Traefik Dashboard",
		"registry-server":      "Docker Registry server",
		"registry-user":        "Docker Registry username",
		"registry-pass":        "Docker Registry password",
		"backup-s3-endpoint":   "S3 endpoint cho backup",
		"backup-s3-bucket":     "S3 bucket name",
		"backup-s3-access-key": "S3 access key",
		"backup-s3-secret-key": "S3 secret key",
		"backup-s3-region":     "S3 region (mặc định: us-east-1)",
		"backup-s3-schedule":   "Cron schedule backup (mặc định: 0 2 * * *)",
		"alert-telegram-token": "Telegram Bot Token",
		"alert-telegram-chat":  "Telegram Chat ID",
		"alert-telegram-enabled": "Bật/tắt Telegram alert: true | false",
	}
}

// ── File I/O ──

var defaultConfigPath = filepath.Join(os.Getenv("HOME"), ".swarm-ctl", "config.yml")

func Load() (*Config, error) {
	return LoadFrom(defaultConfigPath)
}

func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
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

func (c *Config) Save() error {
	return c.SaveTo(defaultConfigPath)
}

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

func (c *Config) AddCluster(cl Cluster) {
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
