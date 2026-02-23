package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_EmptyContext(t *testing.T) {
	// Giả lập đường dẫn Config là 1 thư mục temp rỗng
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	configPath := filepath.Join(tempDir, ".swarm-ctl", "config.yml")
	
	// Khởi chạy Load Config từ file test rỗng
	cfg, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("Lỗi không mong muốn khi tạo Config mặc định: %v", err)
	}

	// Xác nhận giá trị mặc định được tự Generator thành công
	if cfg.Version != "1" {
		t.Errorf("Kỳ vọng: Version='1', thực tế: '%s'", cfg.Version)
	}

	if len(cfg.Clusters) != 0 {
		t.Errorf("Kỳ vọng: Chưa có cluster nào được khởi tạo, thực tế: %d", len(cfg.Clusters))
	}

	// Test lưu file
	cfg.AddCluster(Cluster{
		Name:     "production",
		DataRoot: "/opt/data",
	})
	cfg.SaveTo(configPath)

	// Đảm bảo file cấu hình vật lý đã được tạo thành công
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("File cấu hình chưa được tạo tại: %s", configPath)
	}
}
