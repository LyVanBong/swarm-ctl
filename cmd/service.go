package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/template"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Quản lý services trong cluster",
}

// ──────────────────────────────────────────────
// swarm-ctl service add
// ──────────────────────────────────────────────
var (
	svcAddName        string
	svcAddImage       string
	svcAddDomain      string
	svcAddPort        int
	svcAddReplicas    int
	svcAddCPU         string
	svcAddMemory      string
	svcAddPlacement   string
	svcAddMiddleware  []string
	svcAddEnv         []string
	svcAddSecret      []string
	svcAddVolume      []string
	svcAddNetwork     []string
	svcAddRestart     string
	svcAddUpdateOrder string
)

var serviceAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Thêm service mới vào cluster",
	Long: `Tạo và deploy service mới với wizard hướng dẫn.

Tool tự động:
  - Generate docker-compose.yml với Traefik labels
  - Đăng ký vào services.yml
  - Deploy lên cluster
  - Cấp SSL tự động qua Let's Encrypt

Ví dụ:
  swarm-ctl service add --name my-api --image nginx:latest --domain api.example.com --port 3000
  swarm-ctl service add --name my-api --image nginx --domain api.example.com --port 3000 --replicas 3 --placement tier=app`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate
		if svcAddName == "" || svcAddImage == "" || svcAddDomain == "" || svcAddPort == 0 {
			return fmt.Errorf("thiếu thông tin bắt buộc: --name, --image, --domain, --port")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render(fmt.Sprintf("📦 ADD SERVICE: %s", svcAddName)))
		fmt.Println()

		// Hiển thị config sẽ tạo
		fmt.Println(ui.SectionHeader.Render(" CẤU HÌNH SERVICE "))
		fmt.Printf("  %-15s: %s\n", "Tên", ui.Info.Render(svcAddName))
		fmt.Printf("  %-15s: %s\n", "Image", ui.Info.Render(svcAddImage))
		fmt.Printf("  %-15s: %s\n", "Domain", ui.Info.Render("https://"+svcAddDomain))
		fmt.Printf("  %-15s: %s\n", "Port", ui.Info.Render(fmt.Sprintf("%d", svcAddPort)))
		fmt.Printf("  %-15s: %s\n", "Replicas", ui.Info.Render(fmt.Sprintf("%d", svcAddReplicas)))
		fmt.Printf("  %-15s: %s\n", "CPU Limit", ui.Info.Render(svcAddCPU))
		fmt.Printf("  %-15s: %s\n", "Memory Limit", ui.Info.Render(svcAddMemory))
		if svcAddPlacement != "" {
			fmt.Printf("  %-15s: %s\n", "Placement", ui.Info.Render("label: "+svcAddPlacement))
		}
		if len(svcAddMiddleware) > 0 {
			fmt.Printf("  %-15s: %s\n", "Middleware", ui.Info.Render(strings.Join(svcAddMiddleware, ", ")))
		}
		if len(svcAddSecret) > 0 {
			fmt.Printf("  %-15s: %s\n", "Secrets", ui.Info.Render(strings.Join(svcAddSecret, ", ")))
		}
		fmt.Println()

		// Generate service files
		fmt.Println(ui.RenderStep(1, 3, "Generating service files..."))

		svcTemplate := template.ServiceSpec{
			Name:        svcAddName,
			Image:       svcAddImage,
			Domain:      svcAddDomain,
			Port:        svcAddPort,
			Replicas:    svcAddReplicas,
			CPULimit:    svcAddCPU,
			MemoryLimit: svcAddMemory,
			Placement:   svcAddPlacement,
			Middlewares: svcAddMiddleware,
			Env:         parseKeyValueList(svcAddEnv),
			Secrets:     svcAddSecret,
			Volumes:     svcAddVolume,
			Networks:    svcAddNetwork,
			RestartPolicy: svcAddRestart,
			UpdateOrder: svcAddUpdateOrder,
		}

		// Tạo thư mục service
		serviceDir := fmt.Sprintf("./services/apps/%s", svcAddName)
		if err := os.MkdirAll(serviceDir, 0755); err != nil {
			return fmt.Errorf("không thể tạo thư mục service: %w", err)
		}

		// Generate docker-compose.yml
		composeContent, err := template.GenerateCompose(svcTemplate)
		if err != nil {
			return fmt.Errorf("lỗi generate compose file: %w", err)
		}

		composePath := fmt.Sprintf("%s/docker-compose.yml", serviceDir)
		if err := os.WriteFile(composePath, []byte(composeContent), 0644); err != nil {
			return fmt.Errorf("không thể ghi file: %w", err)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Đã tạo: %s", composePath)))

		// Deploy
		fmt.Println(ui.RenderStep(2, 3, "Deploy lên cluster..."))
		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return fmt.Errorf("không thể kết nối cluster: %w", err)
		}
		defer client.Close()

		// Upload compose file lên master
		if err := client.WriteFile(fmt.Sprintf("/tmp/%s-compose.yml", svcAddName), composeContent); err != nil {
			return fmt.Errorf("upload compose file thất bại: %w", err)
		}

		// Deploy stack
		output, err := client.RunSudo(fmt.Sprintf(
			"docker stack deploy -c /tmp/%s-compose.yml %s --with-registry-auth",
			svcAddName, svcAddName))
		if err != nil {
			return fmt.Errorf("deploy thất bại: %w\n%s", err, output)
		}
		fmt.Println(ui.RenderSuccess("Service đã được deploy"))

		// Verify
		fmt.Println(ui.RenderStep(3, 3, "Kiểm tra service..."))
		verifyOutput, _ := client.Run(fmt.Sprintf(
			"docker service ls --filter 'label=com.docker.stack.namespace=%s'", svcAddName))
		fmt.Println(verifyOutput)

		fmt.Println(ui.SuccessBox.Render(fmt.Sprintf(`
✅ Service "%s" đã chạy thành công!

   🌐 URL  : https://%s
   📦 Image: %s
   🔢 Reps : %d

   Quản lý:
   → swarm-ctl service scale %s=%d
   → swarm-ctl service logs %s
   → swarm-ctl service update %s --image %s:new-tag
`, svcAddName, svcAddDomain, svcAddImage, svcAddReplicas,
			svcAddName, svcAddReplicas, svcAddName, svcAddName, svcAddImage)))

		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl service deploy
// ──────────────────────────────────────────────
var serviceDeployCmd = &cobra.Command{
	Use:   "deploy [service-name]",
	Short: "Deploy service từ services.yml",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.RenderStep(1, 2, fmt.Sprintf("Deploy %s...", serviceName)))

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		output, err := client.Run(fmt.Sprintf("cd /home/products && bash scripts/deploy-smart.sh %s", serviceName))
		if err != nil {
			return fmt.Errorf("deploy thất bại:\n%s", output)
		}
		fmt.Println(output)
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Service %s đã được deploy", serviceName)))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl service scale
// ──────────────────────────────────────────────
var serviceScaleCmd = &cobra.Command{
	Use:   "scale [service=replicas]",
	Short: "Scale service lên/xuống số replicas",
	Long: `Ví dụ:
  swarm-ctl service scale my-api=5
  swarm-ctl service scale appwrite=3 appwrite-worker=10`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		for _, arg := range args {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				fmt.Println(ui.RenderWarning(fmt.Sprintf("Bỏ qua '%s' — format phải là service=replicas", arg)))
				continue
			}

			fmt.Printf("  Scaling %s → %s replicas...\n", ui.Bold.Render(parts[0]), ui.Info.Render(parts[1]))
			output, err := client.Run(fmt.Sprintf("docker service scale %s=%s", parts[0], parts[1]))
			if err != nil {
				fmt.Println(ui.RenderError(fmt.Sprintf("Scale %s thất bại: %s", parts[0], err.Error())))
				continue
			}
			fmt.Println(output)
		}
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl service list
// ──────────────────────────────────────────────
var serviceListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Xem danh sách tất cả services",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		fmt.Println(ui.SectionHeader.Render(fmt.Sprintf(" SERVICES — %s ", cluster.Name)))

		output, err := client.Run(`docker service ls --format "table {{.Name}}\t{{.Mode}}\t{{.Replicas}}\t{{.Image}}\t{{.Ports}}"`)
		if err != nil {
			return err
		}

		lines := strings.Split(strings.TrimSpace(output), "\n")
		for i, line := range lines {
			if i == 0 {
				fmt.Println(ui.Bold.Render(line))
			} else {
				fmt.Println(line)
			}
		}
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl service logs
// ──────────────────────────────────────────────
var serviceLogsCmd = &cobra.Command{
	Use:   "logs [service-name]",
	Short: "Xem logs của service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		fmt.Println(ui.Info.Render(fmt.Sprintf("📋 Logs: %s", args[0])))
		output, err := client.Run(fmt.Sprintf("docker service logs --tail 100 --follow %s", args[0]))
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl service remove
// ──────────────────────────────────────────────
var serviceRemoveCmd = &cobra.Command{
	Use:     "remove [service-name]",
	Short:   "Xóa service khỏi cluster",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.WarnBox.Render(fmt.Sprintf("⚠️  Xóa service: %s\n   (Data volumes sẽ KHÔNG bị xóa)", serviceName)))
		fmt.Print("\nXác nhận? (yes/no): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			fmt.Println(ui.RenderInfo("Đã hủy"))
			return nil
		}

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		output, err := client.Run(fmt.Sprintf("docker stack rm %s", serviceName))
		if err != nil {
			return fmt.Errorf("xóa thất bại: %w\n%s", err, output)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Service %s đã được xóa", serviceName)))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl service rollback
// ──────────────────────────────────────────────
var serviceRollbackCmd = &cobra.Command{
	Use:   "rollback [service-name]",
	Short: "Rollback service về version trước",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		fmt.Printf("🔄 Rollback %s...\n", args[0])
		output, err := client.Run(fmt.Sprintf("docker service rollback %s", args[0]))
		if err != nil {
			return fmt.Errorf("rollback thất bại: %w\n%s", err, output)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Service %s đã được rollback", args[0])))
		return nil
	},
}

func init() {
	// service add flags
	serviceAddCmd.Flags().StringVarP(&svcAddName, "name", "n", "", "Tên service (bắt buộc)")
	serviceAddCmd.Flags().StringVarP(&svcAddImage, "image", "i", "", "Docker image (bắt buộc)")
	serviceAddCmd.Flags().StringVarP(&svcAddDomain, "domain", "d", "", "Domain name (bắt buộc)")
	serviceAddCmd.Flags().IntVarP(&svcAddPort, "port", "p", 0, "Port ứng dụng lắng nghe (bắt buộc)")
	serviceAddCmd.Flags().IntVarP(&svcAddReplicas, "replicas", "r", 1, "Số replicas")
	serviceAddCmd.Flags().StringVar(&svcAddCPU, "cpu", "0.5", "CPU limit (vd: 0.5, 2.0)")
	serviceAddCmd.Flags().StringVar(&svcAddMemory, "memory", "512M", "Memory limit (vd: 512M, 2G)")
	serviceAddCmd.Flags().StringVar(&svcAddPlacement, "placement", "", "Placement constraint label (vd: tier=app)")
	serviceAddCmd.Flags().StringArrayVar(&svcAddMiddleware, "middleware", nil, "Traefik middlewares (vd: ratelimit,security-headers)")
	serviceAddCmd.Flags().StringArrayVarP(&svcAddEnv, "env", "e", nil, "Environment variables (vd: KEY=value)")
	serviceAddCmd.Flags().StringArrayVar(&svcAddSecret, "secret", nil, "Docker secrets cần thiết")
	serviceAddCmd.Flags().StringArrayVar(&svcAddVolume, "volume", nil, "Volume mounts (vd: /host:/container)")
	serviceAddCmd.Flags().StringArrayVar(&svcAddNetwork, "network", nil, "Networks bổ sung")
	serviceAddCmd.Flags().StringVar(&svcAddRestart, "restart", "on-failure", "Restart policy: on-failure|any|none")
	serviceAddCmd.Flags().StringVar(&svcAddUpdateOrder, "update-order", "start-first", "Update order: start-first|stop-first")

	serviceAddCmd.MarkFlagRequired("name")
	serviceAddCmd.MarkFlagRequired("image")
	serviceAddCmd.MarkFlagRequired("domain")
	serviceAddCmd.MarkFlagRequired("port")

	// Register subcommands
	serviceCmd.AddCommand(serviceAddCmd)
	serviceCmd.AddCommand(serviceDeployCmd)
	serviceCmd.AddCommand(serviceScaleCmd)
	serviceCmd.AddCommand(serviceListCmd)
	serviceCmd.AddCommand(serviceLogsCmd)
	serviceCmd.AddCommand(serviceRemoveCmd)
	serviceCmd.AddCommand(serviceRollbackCmd)
}

// parseKeyValueList parse ["KEY=value"] thành map[string]string
func parseKeyValueList(list []string) map[string]string {
	result := make(map[string]string)
	for _, item := range list {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}
