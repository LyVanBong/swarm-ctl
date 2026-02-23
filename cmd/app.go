package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Quản lý Ứng dụng (App Bundle) trên Swarm",
	Long:  `Triển khai và Cấu hình Ứng dụng từ Thư mục (Bundle) lên Cụm Docker Swarm.`,
}

var appName string
var composeFiles []string

var appDeployCmd = &cobra.Command{
	Use:   "deploy [FOLDER_PATH]",
	Short: "Triển khai một ứng dụng Bundle (có docker-compose.yml)",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		folderPath := args[0]
		absPath, err := filepath.Abs(folderPath)
		if err != nil {
			return fmt.Errorf("đường dẫn không hợp lệ: %w", err)
		}

		info, err := os.Stat(absPath)
		if err != nil || !info.IsDir() {
			return fmt.Errorf("đường dẫn '%s' không tồn tại hoặc không phải là thư mục", absPath)
		}

		// Xác định App Name
		if appName == "" {
			appName = filepath.Base(absPath)
			appName = strings.ToLower(appName)
		}

		// Tải cấu hình cluster (Env)
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render(fmt.Sprintf("🚀 DEPLOYING APP BUNDLE: %s", appName)))
		fmt.Printf("  Thư mục gốc: %s\n\n", ui.Info.Render(absPath))

		// Kết nối SSH
		fmt.Println(ui.RenderStep(1, 3, "Chuẩn bị nén Bundle (YML + Configs + Secrets)..."))
		tarFile := fmt.Sprintf("/tmp/%s-bundle.tar.gz", appName)
		tarCmd := exec.Command("tar", "-czf", tarFile, "-C", absPath, ".")
		if err := tarCmd.Run(); err != nil {
			return fmt.Errorf("lỗi nén thư mục bundle: %w", err)
		}
		defer os.Remove(tarFile)

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		fmt.Println(ui.RenderStep(2, 3, "Bắn gói Bundle lên Server Master..."))
		remoteDir := fmt.Sprintf("/opt/swarm-ctl-apps/%s", appName)
		client.Run(fmt.Sprintf("mkdir -p %s", remoteDir))

		// Tải file qua thư viện go-scp (Bao bọc đa nền tảng)
		scpClient, err := scp.NewClientBySSH(client.GetRawClient())
		if err != nil {
			return fmt.Errorf("không thể khởi tạo giao thức scp: %w", err)
		}
		defer scpClient.Close()
		
		fTar, err := os.Open(tarFile)
		if err != nil {
			return fmt.Errorf("lỗi đọc file nén: %w", err)
		}
		defer fTar.Close()

		if err := scpClient.CopyFromFile(cmd.Context(), *fTar, remoteDir+"/bundle.tar.gz", "0644"); err != nil {
			return fmt.Errorf("lỗi trong quá trình truyền file Bundle: %w", err)
		}

		fmt.Println(ui.RenderStep(3, 3, "Triển khai Docker Stack Native..."))
		
		envInjection := fmt.Sprintf("export DATA_ROOT='%s' && export DOMAIN='%s' && ", cluster.DataRoot, cluster.Domain)

		composeArgs := "-c docker-compose.yml"
		for _, f := range composeFiles {
			composeArgs += fmt.Sprintf(" -c %s", f)
		}

		deployScript := fmt.Sprintf(`cd %s && tar -xzf bundle.tar.gz && rm bundle.tar.gz && %s docker stack deploy --with-registry-auth %s %s`,
			remoteDir, envInjection, composeArgs, appName)

		// Extract and create volumes
		composeData, err := os.ReadFile(filepath.Join(absPath, "docker-compose.yml"))
		if err == nil {
			type Compose struct {
				Volumes map[string]struct {
					DriverOpts map[string]string `yaml:"driver_opts"`
				} `yaml:"volumes"`
			}
			var c Compose
			if err := yaml.Unmarshal(composeData, &c); err == nil {
				for _, v := range c.Volumes {
					if v.DriverOpts != nil {
						device := v.DriverOpts["device"]
						if device != "" {
							// Replace DATA_ROOT variable
							parsedPath := strings.ReplaceAll(device, "${DATA_ROOT}", cluster.DataRoot)
							parsedPath = strings.ReplaceAll(parsedPath, "$DATA_ROOT", cluster.DataRoot)
							if strings.HasPrefix(parsedPath, "/") {
								fmt.Printf("📂 Tạo thư mục Local Volume: %s\n", parsedPath)
								client.Run(fmt.Sprintf("mkdir -p %s", parsedPath))
							}
						}
					}
				}
			}
		}

		output, err := client.Run(deployScript)
		if err != nil {
			return fmt.Errorf("triển khai stack thất bại: %w\nOutput: %s", err, output)
		}

		fmt.Println()
		fmt.Println(ui.SuccessBox.Render(fmt.Sprintf(`
✅ Ứng dụng '%s' đã được yêu cầu triển khai!

    Docker Swarm đang phân bổ tài nguyên.
    Vui lòng kiểm tra logs để xem chi tiết khởi động:
      swarm-ctl service logs %s_web
      swarm-ctl service ls
`, appName, appName)))

		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl app remove
// ──────────────────────────────────────────────
var appRemoveCmd = &cobra.Command{
	Use:     "remove [APP_NAME]",
	Aliases: []string{"rm"},
	Short:   "Gỡ bỏ hoàn toàn một ứng dụng khỏi cụm Server",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render(fmt.Sprintf("🗑️  REMOVING APP: %s", appName)))

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		fmt.Printf("Đang yêu cầu hệ thống tháo rỡ Service '%s' trên toàn bộ các Server...\n", appName)

		output, err := client.Run(fmt.Sprintf("docker stack rm %s", appName))
		if err != nil {
			return fmt.Errorf("không thể gỡ ứng dụng %s: %w\n%s", appName, err, output)
		}

		remoteDir := fmt.Sprintf("/opt/swarm-ctl-apps/%s", appName)
		client.Run(fmt.Sprintf("rm -rf %s", remoteDir))

		fmt.Println()
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Toàn bộ Container của '%s' đã ngắt điện trơn tru!", appName)))

		return nil
	},
}

func init() {
	appDeployCmd.Flags().StringVarP(&appName, "name", "n", "", "Tên Stack App (Vd: webapp, nginx)")
	appDeployCmd.Flags().StringSliceVarP(&composeFiles, "compose-file", "c", []string{}, "Sử dụng các file compose override (Vd: -c docker-compose.prod.yml)")

	appCmd.AddCommand(appDeployCmd)
	appCmd.AddCommand(appRemoveCmd)
}
