package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Quản lý cấu hình cluster (domain, subdomain, SSL, auth, registry, backup, alert...)",
	Long: `Xem và cập nhật cấu hình cluster hiện tại.

Ví dụ:
  swarm-ctl config show                                # Xem toàn bộ cấu hình
  swarm-ctl config set domain company.com              # Đổi domain chính
  swarm-ctl config set traefik-subdomain dashboard     # Đổi subdomain Traefik
  swarm-ctl config set portainer-subdomain admin       # Đổi subdomain Portainer
  swarm-ctl config set acme-email devops@company.com   # Đổi email SSL
  swarm-ctl config set acme-challenge http             # Đổi kiểu xác thực SSL
  swarm-ctl config reset traefik-subdomain             # Đặt lại về mặc định
  swarm-ctl config apply                               # Áp dụng lên server đang chạy
  swarm-ctl config export config-backup.yml            # Xuất config ra file
  swarm-ctl config import config-backup.yml            # Nhập config từ file
  swarm-ctl config diff                                # So sánh config local vs server
  swarm-ctl config keys                                # Xem danh sách tất cả key hỗ trợ`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// ──────────────────────────────────────────────
// swarm-ctl config show
// ──────────────────────────────────────────────
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Xem cấu hình cluster hiện tại",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render("⚙️  CLUSTER CONFIG"))
		fmt.Println()

		// ── Thông tin chung ──
		fmt.Println(ui.SectionHeader.Render(" THÔNG TIN CHUNG "))
		fmt.Printf("  Cluster Name : %s\n", ui.Info.Render(cluster.Name))
		fmt.Printf("  Master IP    : %s\n", ui.Info.Render(cluster.MasterIP))
		fmt.Printf("  Domain       : %s\n", ui.Info.Render(cluster.Domain))
		fmt.Printf("  SSH User     : %s\n", ui.Info.Render(cluster.SSHUser))
		fmt.Printf("  SSH Key      : %s\n", ui.Info.Render(cluster.SSHKey))
		fmt.Printf("  Data Root    : %s\n", ui.Info.Render(cluster.DataRoot))
		fmt.Printf("  Created At   : %s\n", ui.Info.Render(cluster.CreatedAt))
		fmt.Println()

		// ── Subdomains ──
		fmt.Println(ui.SectionHeader.Render(" SUBDOMAINS "))
		fmt.Printf("  🌐 Traefik   : %s\n", ui.Success.Render(cluster.GetTraefikHost()))
		fmt.Printf("  📊 Portainer : %s\n", ui.Success.Render(cluster.GetPortainerHost()))
		fmt.Println()

		// ── SSL / ACME ──
		fmt.Println(ui.SectionHeader.Render(" SSL / ACME "))
		fmt.Printf("  Email        : %s\n", ui.Info.Render(cluster.GetACMEEmail()))
		fmt.Printf("  Challenge    : %s\n", ui.Info.Render(cluster.GetACMEChallenge()))
		fmt.Println()

		// ── Traefik Auth ──
		fmt.Println(ui.SectionHeader.Render(" TRAEFIK AUTH "))
		fmt.Printf("  Username     : %s\n", ui.Info.Render(cluster.GetAuthUsername()))
		passDisplay := "(không đặt)"
		if cluster.Auth.Password != "" {
			passDisplay = "********"
		}
		fmt.Printf("  Password     : %s\n", ui.Info.Render(passDisplay))
		fmt.Println()

		// ── Docker Registry ──
		fmt.Println(ui.SectionHeader.Render(" DOCKER REGISTRY "))
		fmt.Printf("  Server       : %s\n", ui.Info.Render(cluster.GetRegistryServer()))
		fmt.Printf("  Username     : %s\n", ui.Info.Render(valueOrDefault(cluster.Registry.Username, "(không đặt)")))
		regPassDisplay := "(không đặt)"
		if cluster.Registry.Password != "" {
			regPassDisplay = "********"
		}
		fmt.Printf("  Password     : %s\n", ui.Info.Render(regPassDisplay))
		fmt.Println()

		// ── Backup S3 ──
		fmt.Println(ui.SectionHeader.Render(" BACKUP S3 "))
		fmt.Printf("  Endpoint     : %s\n", ui.Info.Render(valueOrDefault(cluster.Backup.Endpoint, "(không đặt)")))
		fmt.Printf("  Bucket       : %s\n", ui.Info.Render(valueOrDefault(cluster.Backup.Bucket, "(không đặt)")))
		fmt.Printf("  Region       : %s\n", ui.Info.Render(cluster.GetBackupRegion()))
		fmt.Printf("  Schedule     : %s\n", ui.Info.Render(cluster.GetBackupSchedule()))
		s3KeyDisplay := "(không đặt)"
		if cluster.Backup.AccessKey != "" {
			s3KeyDisplay = "********"
		}
		fmt.Printf("  Access Key   : %s\n", ui.Info.Render(s3KeyDisplay))
		fmt.Println()

		// ── Telegram Alert ──
		fmt.Println(ui.SectionHeader.Render(" TELEGRAM ALERT "))
		alertStatus := "❌ Tắt"
		if cluster.Alert.Enabled {
			alertStatus = "✅ Bật"
		}
		fmt.Printf("  Trạng thái   : %s\n", ui.Info.Render(alertStatus))
		fmt.Printf("  Chat ID      : %s\n", ui.Info.Render(valueOrDefault(cluster.Alert.ChatID, "(không đặt)")))
		tokenDisplay := "(không đặt)"
		if cluster.Alert.BotToken != "" {
			tokenDisplay = "********"
		}
		fmt.Printf("  Bot Token    : %s\n", ui.Info.Render(tokenDisplay))
		fmt.Println()

		fmt.Println(ui.RenderInfo("Thay đổi: swarm-ctl config set <key> <value>"))
		fmt.Println(ui.RenderInfo("Áp dụng : swarm-ctl config apply"))
		fmt.Println(ui.RenderInfo("Xem key : swarm-ctl config keys"))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl config keys
// ──────────────────────────────────────────────
var configKeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Liệt kê tất cả config keys được hỗ trợ",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.Banner.Render("🔑 CONFIG KEYS"))
		fmt.Println()

		allKeys := config.AllConfigKeys()

		// Nhóm theo prefix
		groups := map[string][]string{
			"🏗️  Cluster":   {"domain", "cluster-name", "data-root", "ssh-user", "ssh-key"},
			"🌐 Subdomains": {"traefik-subdomain", "portainer-subdomain"},
			"🔒 SSL/ACME":   {"acme-email", "acme-challenge"},
			"🛡️  Auth":      {"traefik-auth-user", "traefik-auth-pass"},
			"📦 Registry":   {"registry-server", "registry-user", "registry-pass"},
			"💾 Backup S3":  {"backup-s3-endpoint", "backup-s3-bucket", "backup-s3-access-key", "backup-s3-secret-key", "backup-s3-region", "backup-s3-schedule"},
			"🔔 Telegram":   {"alert-telegram-token", "alert-telegram-chat", "alert-telegram-enabled"},
		}

		order := []string{"🏗️  Cluster", "🌐 Subdomains", "🔒 SSL/ACME", "🛡️  Auth", "📦 Registry", "💾 Backup S3", "🔔 Telegram"}

		for _, group := range order {
			keys := groups[group]
			fmt.Println(ui.SectionHeader.Render(fmt.Sprintf(" %s ", group)))
			for _, k := range keys {
				desc := allKeys[k]
				fmt.Printf("  %-28s %s\n", ui.Info.Render(k), desc)
			}
			fmt.Println()
		}

		fmt.Println(ui.RenderInfo("Sử dụng: swarm-ctl config set <key> <value>"))
	},
}

// ──────────────────────────────────────────────
// swarm-ctl config set <key> <value>
// ──────────────────────────────────────────────
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Cập nhật giá trị cấu hình",
	Long: `Cập nhật giá trị cấu hình cho cluster hiện tại.

Xem danh sách key: swarm-ctl config keys

Ví dụ:
  swarm-ctl config set domain company.com
  swarm-ctl config set traefik-subdomain dashboard
  swarm-ctl config set acme-challenge http
  swarm-ctl config set alert-telegram-enabled true`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])
		value := args[1]

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		// Validate key
		allKeys := config.AllConfigKeys()
		if _, ok := allKeys[key]; !ok {
			return fmt.Errorf("key không hợp lệ: '%s'\n\nXem danh sách key: swarm-ctl config keys", key)
		}

		oldValue := applyConfigValue(cluster, key, value)

		cfg.UpdateCluster(*cluster)
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("lưu config thất bại: %w", err)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("[%s]: %s → %s", key, oldValue, value)))
		fmt.Println()
		fmt.Println(ui.RenderWarning("Chạy 'swarm-ctl config apply' để áp dụng thay đổi lên server đang chạy"))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl config reset <key>
// ──────────────────────────────────────────────
var configResetCmd = &cobra.Command{
	Use:   "reset <key>",
	Short: "Đặt lại giá trị về mặc định",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		allKeys := config.AllConfigKeys()
		if _, ok := allKeys[key]; !ok {
			return fmt.Errorf("key không hợp lệ: '%s'", key)
		}

		// Reset = đặt về chuỗi rỗng (helper sẽ trả về giá trị mặc định)
		applyConfigValue(cluster, key, "")

		cfg.UpdateCluster(*cluster)
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("lưu config thất bại: %w", err)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("[%s] đã được đặt lại về giá trị mặc định", key)))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl config apply
// ──────────────────────────────────────────────
var configApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Áp dụng cấu hình mới lên server đang chạy",
	Long: `Đọc cấu hình từ file config và cập nhật Traefik/Portainer trên server.

Lệnh này sẽ:
  1. Cập nhật routing subdomain cho Traefik Dashboard
  2. Cập nhật routing subdomain cho Portainer
  3. Xác minh services đang chạy

Ví dụ:
  swarm-ctl config set traefik-subdomain dashboard
  swarm-ctl config set portainer-subdomain admin
  swarm-ctl config apply`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		traefikHost := cluster.GetTraefikHost()
		portainerHost := cluster.GetPortainerHost()

		fmt.Println(ui.Banner.Render("🔄 CONFIG APPLY"))
		fmt.Println()
		fmt.Printf("  🌐 Traefik   → %s\n", ui.Success.Render(traefikHost))
		fmt.Printf("  📊 Portainer → %s\n", ui.Success.Render(portainerHost))
		fmt.Println()

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return fmt.Errorf("không thể kết nối tới master: %w", err)
		}
		defer client.Close()

		// Step 1: Cập nhật Traefik labels
		fmt.Println(ui.RenderStep(1, 3, "Cập nhật routing Traefik Dashboard..."))
		traefikCmd := fmt.Sprintf(
			`docker service update `+
				`--label-add "traefik.http.routers.traefik.rule=Host(%s)" `+
				`traefik_traefik --quiet 2>&1`,
			"`"+traefikHost+"`")
		if _, err := client.Run(traefikCmd); err != nil {
			fmt.Println(ui.RenderWarning(fmt.Sprintf("Traefik update warning: %v", err)))
		} else {
			fmt.Println(ui.RenderSuccess(fmt.Sprintf("Traefik → https://%s", traefikHost)))
		}

		// Step 2: Cập nhật Portainer labels
		fmt.Println(ui.RenderStep(2, 3, "Cập nhật routing Portainer..."))
		portainerCmd := fmt.Sprintf(
			`docker service update `+
				`--label-add "traefik.http.routers.portainer.rule=Host(%s)" `+
				`portainer_portainer --quiet 2>&1`,
			"`"+portainerHost+"`")
		if _, err := client.Run(portainerCmd); err != nil {
			fmt.Println(ui.RenderWarning(fmt.Sprintf("Portainer update warning: %v", err)))
		} else {
			fmt.Println(ui.RenderSuccess(fmt.Sprintf("Portainer → https://%s", portainerHost)))
		}

		// Step 3: Xác minh
		fmt.Println(ui.RenderStep(3, 3, "Xác minh services..."))
		output, _ := client.Run("docker service ls --format '{{.Name}}\\t{{.Replicas}}'")
		if output != "" {
			fmt.Println(output)
		}

		fmt.Println()
		fmt.Println(ui.SuccessBox.Render(fmt.Sprintf(`
✅ Cấu hình đã được áp dụng thành công!

  🌐 Traefik Dashboard : https://%s
  📊 Portainer         : https://%s

  Lưu ý: Hãy đảm bảo DNS đã trỏ subdomain mới về IP: %s
`, traefikHost, portainerHost, cluster.MasterIP)))

		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl config export <file>
// ──────────────────────────────────────────────
var configExportCmd = &cobra.Command{
	Use:   "export [file]",
	Short: "Xuất cấu hình ra file (YAML hoặc JSON)",
	Long: `Xuất toàn bộ cấu hình cluster hiện tại ra file.

Ví dụ:
  swarm-ctl config export                       # In ra stdout (YAML)
  swarm-ctl config export backup.yml            # Lưu file YAML
  swarm-ctl config export backup.json           # Lưu file JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		var data []byte
		outputFile := ""
		if len(args) > 0 {
			outputFile = args[0]
		}

		if outputFile != "" && strings.HasSuffix(outputFile, ".json") {
			data, err = json.MarshalIndent(cluster, "", "  ")
		} else {
			data, err = yaml.Marshal(cluster)
		}
		if err != nil {
			return fmt.Errorf("serialize thất bại: %w", err)
		}

		if outputFile == "" {
			fmt.Println(string(data))
			return nil
		}

		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			return fmt.Errorf("ghi file thất bại: %w", err)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Config đã được xuất ra: %s", outputFile)))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl config import <file>
// ──────────────────────────────────────────────
var configImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Nhập cấu hình từ file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		importFile := args[0]

		data, err := os.ReadFile(importFile)
		if err != nil {
			return fmt.Errorf("không thể đọc file: %w", err)
		}

		var imported config.Cluster

		if strings.HasSuffix(importFile, ".json") {
			err = json.Unmarshal(data, &imported)
		} else {
			err = yaml.Unmarshal(data, &imported)
		}
		if err != nil {
			return fmt.Errorf("parse file thất bại: %w", err)
		}

		if imported.Name == "" || imported.MasterIP == "" {
			return fmt.Errorf("file config không hợp lệ: thiếu 'name' hoặc 'master_ip'")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		cfg.AddCluster(imported)
		cfg.Current = imported.Name
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("lưu config thất bại: %w", err)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Đã import config cluster '%s' từ %s", imported.Name, importFile)))
		fmt.Println(ui.RenderInfo("Chạy 'swarm-ctl config show' để xem chi tiết"))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl config diff
// ──────────────────────────────────────────────
var configDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "So sánh config local với cấu hình đang chạy trên server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render("🔍 CONFIG DIFF"))
		fmt.Println()

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return fmt.Errorf("không thể kết nối tới master: %w", err)
		}
		defer client.Close()

		hasDiff := false

		// Kiểm tra Traefik routing rule trên server
		traefikRule, _ := client.Run(
			`docker service inspect traefik_traefik --format '{{range .Spec.Labels}}{{.}}{{end}}' 2>/dev/null | grep -oP 'Host\(\x60\K[^\x60]+'`)
		traefikRule = strings.TrimSpace(traefikRule)
		localTraefik := cluster.GetTraefikHost()

		fmt.Println(ui.SectionHeader.Render(" TRAEFIK DASHBOARD "))
		if traefikRule != "" && traefikRule != localTraefik {
			fmt.Printf("  Server : %s\n", ui.Warning.Render(traefikRule))
			fmt.Printf("  Local  : %s\n", ui.Info.Render(localTraefik))
			fmt.Println(ui.RenderWarning("⚠️  KHÁC BIỆT — Chạy 'config apply' để đồng bộ"))
			hasDiff = true
		} else {
			fmt.Printf("  ✅ Đồng bộ: %s\n", ui.Success.Render(localTraefik))
		}
		fmt.Println()

		// Kiểm tra Portainer routing rule trên server
		portainerRule, _ := client.Run(
			`docker service inspect portainer_portainer --format '{{range .Spec.Labels}}{{.}}{{end}}' 2>/dev/null | grep -oP 'Host\(\x60\K[^\x60]+'`)
		portainerRule = strings.TrimSpace(portainerRule)
		localPortainer := cluster.GetPortainerHost()

		fmt.Println(ui.SectionHeader.Render(" PORTAINER "))
		if portainerRule != "" && portainerRule != localPortainer {
			fmt.Printf("  Server : %s\n", ui.Warning.Render(portainerRule))
			fmt.Printf("  Local  : %s\n", ui.Info.Render(localPortainer))
			fmt.Println(ui.RenderWarning("⚠️  KHÁC BIỆT — Chạy 'config apply' để đồng bộ"))
			hasDiff = true
		} else {
			fmt.Printf("  ✅ Đồng bộ: %s\n", ui.Success.Render(localPortainer))
		}
		fmt.Println()

		// Kiểm tra Portainer version
		portainerVersion, _ := client.Run(
			`docker service inspect portainer_portainer --format '{{.Spec.TaskTemplate.ContainerSpec.Image}}' 2>/dev/null`)
		portainerVersion = strings.TrimSpace(portainerVersion)
		if portainerVersion != "" {
			fmt.Println(ui.SectionHeader.Render(" VERSIONS "))
			fmt.Printf("  Portainer : %s\n", ui.Info.Render(portainerVersion))

			traefikVersion, _ := client.Run(
				`docker service inspect traefik_traefik --format '{{.Spec.TaskTemplate.ContainerSpec.Image}}' 2>/dev/null`)
			traefikVersion = strings.TrimSpace(traefikVersion)
			fmt.Printf("  Traefik   : %s\n", ui.Info.Render(traefikVersion))
			fmt.Println()
		}

		if !hasDiff {
			fmt.Println(ui.RenderSuccess("✅ Config local và server đang đồng bộ hoàn hảo"))
		}

		return nil
	},
}

// ── Helper functions ──

func valueOrDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

// applyConfigValue gán giá trị mới và trả về giá trị cũ
func applyConfigValue(cluster *config.Cluster, key, value string) string {
	old := ""
	switch key {
	case "domain":
		old = cluster.Domain
		cluster.Domain = value
	case "cluster-name":
		old = cluster.Name
		cluster.Name = value
	case "data-root":
		old = cluster.DataRoot
		cluster.DataRoot = value
	case "ssh-user":
		old = cluster.SSHUser
		cluster.SSHUser = value
	case "ssh-key":
		old = cluster.SSHKey
		cluster.SSHKey = value
	case "traefik-subdomain":
		old = cluster.Subdomains.Traefik
		if old == "" {
			old = "traefik"
		}
		cluster.Subdomains.Traefik = value
	case "portainer-subdomain":
		old = cluster.Subdomains.Portainer
		if old == "" {
			old = "portainer"
		}
		cluster.Subdomains.Portainer = value
	case "acme-email":
		old = cluster.ACME.Email
		cluster.ACME.Email = value
	case "acme-challenge":
		old = cluster.ACME.Challenge
		if old == "" {
			old = "tls"
		}
		cluster.ACME.Challenge = value
	case "traefik-auth-user":
		old = cluster.Auth.Username
		if old == "" {
			old = "admin"
		}
		cluster.Auth.Username = value
	case "traefik-auth-pass":
		old = "********"
		cluster.Auth.Password = value
	case "registry-server":
		old = cluster.Registry.Server
		if old == "" {
			old = "docker.io"
		}
		cluster.Registry.Server = value
	case "registry-user":
		old = cluster.Registry.Username
		cluster.Registry.Username = value
	case "registry-pass":
		old = "********"
		cluster.Registry.Password = value
	case "backup-s3-endpoint":
		old = cluster.Backup.Endpoint
		cluster.Backup.Endpoint = value
	case "backup-s3-bucket":
		old = cluster.Backup.Bucket
		cluster.Backup.Bucket = value
	case "backup-s3-access-key":
		old = "********"
		cluster.Backup.AccessKey = value
	case "backup-s3-secret-key":
		old = "********"
		cluster.Backup.SecretKey = value
	case "backup-s3-region":
		old = cluster.Backup.Region
		if old == "" {
			old = "us-east-1"
		}
		cluster.Backup.Region = value
	case "backup-s3-schedule":
		old = cluster.Backup.Schedule
		if old == "" {
			old = "0 2 * * *"
		}
		cluster.Backup.Schedule = value
	case "alert-telegram-token":
		old = "********"
		cluster.Alert.BotToken = value
	case "alert-telegram-chat":
		old = cluster.Alert.ChatID
		cluster.Alert.ChatID = value
	case "alert-telegram-enabled":
		if cluster.Alert.Enabled {
			old = "true"
		} else {
			old = "false"
		}
		cluster.Alert.Enabled = strings.ToLower(value) == "true"
	}
	return old
}

// Khai báo biến phụ cho sort
var _ = sort.Strings

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configApplyCmd)
	configCmd.AddCommand(configExportCmd)
	configCmd.AddCommand(configImportCmd)
	configCmd.AddCommand(configDiffCmd)
	configCmd.AddCommand(configKeysCmd)
}
