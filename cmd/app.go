package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/LyVanBong/swarm-ctl/internal/catalog"
	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "1-Click Apps (Marketplace/Catolog)",
	Long:  `Cài đặt nhanh các mã nguồn mở thiết yếu (Services phổ biến) vào Swarm.`,
}

var appListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Danh sách các ứng dụng có thể cài đặt 1 chạm",
	Run: func(cmd *cobra.Command, args []string) {
		apps := catalog.GetCatalog()

		fmt.Println(ui.SectionHeader.Render(" MARKETPLACE (1-CLICK APPS) "))
		fmt.Println()

		for _, app := range apps {
			fmt.Printf("📦 %s %s\n", ui.Bold.Render(app.ID), ui.Muted.Render("("+app.Name+")"))
			fmt.Printf("   Chuyên mục : %s\n", ui.Info.Render(app.Category))
			fmt.Printf("   Mô tả      : %s\n", app.Description)
			fmt.Println()
		}

		fmt.Println(ui.RenderInfo("Để cài ứng dụng, chạy: swarm-ctl app install [APP_ID]"))
	},
}

var appDomain string
var appNode string

var appInstallCmd = &cobra.Command{
	Use:   "install [APP_ID]",
	Short: "Triển khai (Deploy) một ứng dụng từ Catalog",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appID := args[0]
		
		// Tìm app trong List
		var selected *catalog.App
		for _, a := range catalog.GetCatalog() {
			if a.ID == appID {
				selected = &a
				break
			}
		}

		if selected == nil {
			return fmt.Errorf("không tìm thấy App ID '%s'. Dùng 'swarm-ctl app list' để xem", appID)
		}

		domain := appDomain
		if domain == "" {
			fmt.Printf("Nhập tên miền (Domain) để gắn cho %s\n", ui.Bold.Render(selected.Name))
			fmt.Print("👉 Domain (vd: n8n.congty.com): ")
			fmt.Scanln(&domain)
			domain = strings.TrimSpace(domain)
			if domain == "" {
				return fmt.Errorf("tên miền không được để trống")
			}
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render(fmt.Sprintf("🚀 INSTALLING: %s", selected.Name)))
		fmt.Printf("  Tên miền: %s\n\n", ui.Info.Render(domain))

		// Render YAML
		appCfg := catalog.AppConfig{
			Domain: domain,
			Node:   appNode,
		}
		yamlStr, err := catalog.GenerateYaml(appID, appCfg)
		if err != nil {
			return fmt.Errorf("lỗi sinh cấu hình compose: %w", err)
		}

		// Tạo file tạm để pipe qua SSH
		tmpFile, err := ioutil.TempFile("", appID+"-*.yml")
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(yamlStr); err != nil {
			return err
		}
		tmpFile.Close()

		// Thực thi deploy
		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		fmt.Println(ui.RenderStep(1, 1, "Gửi cấu hình lên Cluster & Khởi chạy..."))
		
		// Lấy code chuyển nội dung file lên Remote
		composeContent, _ := ioutil.ReadFile(tmpFile.Name())
		remoteDir := fmt.Sprintf("/opt/swarm-ctl-apps/%s", appID)
		
		client.Run("mkdir -p " + remoteDir)
		
		output, err := client.Run(fmt.Sprintf(
			"printf '%%s' '%s' > %s/docker-compose.yml && docker stack deploy -c %s/docker-compose.yml %s",
			strings.ReplaceAll(string(composeContent), "'", "'\\''"),
			remoteDir, remoteDir, appID))

		if err != nil {
			return fmt.Errorf("triển khai stack thất bại: %w\n%s", err, output)
		}

		fmt.Println()
		fmt.Println(ui.SuccessBox.Render(fmt.Sprintf(`
✅ Ứng dụng %s đang được khởi chạy!

   URL Dashboard: https://%s
   
   Hệ thống có thể mất vài phút để sinh chứng chỉ SSL và pull image.
   Xem logs: swarm-ctl service logs %s_%s
`, selected.Name, domain, appID, appID)))

		return nil
	},
}

func init() {
	appInstallCmd.Flags().StringVarP(&appDomain, "domain", "d", "", "Tên miền gắn cho ứng dụng")
	appInstallCmd.Flags().StringVarP(&appNode, "node", "n", "", "Chỉ định Node (Hostname) cụ thể để chạy App")
	
	appCmd.AddCommand(appListCmd)
	appCmd.AddCommand(appInstallCmd)
}
