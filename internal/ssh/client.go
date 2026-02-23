package ssh

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

// Client là SSH client wrapper
type Client struct {
	Host    string
	User    string
	KeyPath string
	Port    int
	client  *gossh.Client
}

// NewClient tạo SSH client mới
func NewClient(host, user, keyPath string) *Client {
	return &Client{
		Host:    host,
		User:    user,
		KeyPath: keyPath,
		Port:    22,
	}
}

// Connect thiết lập SSH connection
func (c *Client) Connect() error {
	key, err := loadPrivateKey(c.KeyPath)
	if err != nil {
		return fmt.Errorf("cannot load SSH key '%s': %w", c.KeyPath, err)
	}

	config := &gossh.ClientConfig{
		User:            c.User,
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(key)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), // TODO: use known_hosts in production
		Timeout:         30 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	client, err := gossh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("cannot connect to %s: %w", addr, err)
	}

	c.client = client
	return nil
}

// Close đóng SSH connection
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// GetRawClient lấy con trỏ gossh client gốc
func (c *Client) GetRawClient() *gossh.Client {
	return c.client
}

// Run chạy lệnh trên remote host, trả về output
func (c *Client) Run(command string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("cannot create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}
	return string(output), nil
}

// RunStream chạy lệnh và stream output realtime
func (c *Client) RunStream(command string, stdout, stderr io.Writer) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("cannot create session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	return session.Run(command)
}

// RunSudo chạy lệnh với sudo
func (c *Client) RunSudo(command string) (string, error) {
	return c.Run("sudo " + command)
}

// FileExists kiểm tra file/thư mục có tồn tại không
func (c *Client) FileExists(path string) (bool, error) {
	output, err := c.Run(fmt.Sprintf("test -e %s && echo 'yes' || echo 'no'", path))
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) == "yes", nil
}

// MkdirAll tạo thư mục trên remote host
func (c *Client) MkdirAll(path string) error {
	_, err := c.RunSudo(fmt.Sprintf("mkdir -p %s", path))
	return err
}

// WriteFile ghi nội dung vào file trên remote host
// Dùng base64 encoding để tránh escape issues
func (c *Client) WriteFile(path, content string) error {
	import64 := "echo '" + encodeBase64(content) + "' | base64 -d | sudo tee " + path + " > /dev/null"
	_, err := c.Run(import64)
	return err
}

// encodeBase64 encode string sang base64
func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// GetHostname lấy hostname của remote host
func (c *Client) GetHostname() (string, error) {
	output, err := c.Run("hostname")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// CheckConnectivity kiểm tra SSH connection nhanh
func CheckConnectivity(host, user, keyPath string) error {
	// Thử TCP connect trước
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:22", host), 5*time.Second)
	if err != nil {
		return fmt.Errorf("port 22 không mở trên %s — kiểm tra firewall", host)
	}
	conn.Close()

	// Thử SSH auth
	client := NewClient(host, user, keyPath)
	if err := client.Connect(); err != nil {
		return err
	}
	defer client.Close()
	return nil
}

// loadPrivateKey đọc SSH private key từ file
func loadPrivateKey(keyPath string) (gossh.Signer, error) {
	// Expand ~ nếu có
	if strings.HasPrefix(keyPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		keyPath = filepath.Join(home, keyPath[2:])
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read key file: %w", err)
	}

	signer, err := gossh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("invalid SSH private key: %w", err)
	}
	return signer, nil
}

// ExpandKeyPath chuyển đổi đường dẫn chứa dâu ~ thành Absolute Path
func ExpandKeyPath(keyPath string) (string, error) {
	if strings.HasPrefix(keyPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, keyPath[2:]), nil
	}
	return keyPath, nil
}

// EnsureSSHKeyExists kiểm tra xem file SSH key có tồn tại hay không. Nếu không, nó sẽ tự động sinh mã khóa Ed25519.
func EnsureSSHKeyExists(keyPath string) (string, error) {
	expandedPath, err := ExpandKeyPath(keyPath)
	if err != nil {
		return keyPath, err
	}

	// Kiểm tra xem File đã tồn tại hay chưa
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		// Đảm bảo là thư mục cha chứa File đã tạo ra
		dir := filepath.Dir(expandedPath)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return keyPath, fmt.Errorf("không thể tạo thư mục %s: %w", dir, err)
		}

		fmt.Printf("Mã khóa SSH chưa tồn tại tại: %s\n", keyPath)
		fmt.Println("Đang tự động sinh khóa Ed25519 bảo mật cao...")

		// Chạy lệnh ssh-keygen
		cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", expandedPath, "-N", "", "-C", "swarm-ctl-auto")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return keyPath, fmt.Errorf("lỗi sinh SSH key: %w\n%s", err, output)
		}
		fmt.Println("Đã sinh SSH Key Ed25519 thành công!")
	}
	
	return expandedPath, nil
}
