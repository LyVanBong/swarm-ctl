package cmd

import (
	"fmt"

	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/spf13/cobra"
)

var (
	registryServer string
	registryUser   string
	registryPass   string
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Quản lý phiên đăng nhập Private Container Registry (Docker Hub, GHCR, v.v.)",
	Long:  "Đăng nhập/Đăng xuất khỏi các Container Registry bảo mật để có thể kéo (pull) Private Images.",
}

var registryLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Đăng nhập vào một Private Container Registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render("🔐 DOCKER REGISTRY LOGIN"))

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		fmt.Printf("Đang yêu cầu Master Node (%s) đăng nhập vào %s...\n", cluster.MasterIP, registryServer)

		// Dùng cờ --password-stdin để bảo mật, không lộ pass ra Process List
		loginCmd := fmt.Sprintf("echo '%s' | docker login %s -u %s --password-stdin", registryPass, registryServer, registryUser)
		output, err := client.Run(loginCmd)
		if err != nil {
			return fmt.Errorf("đăng nhập thất bại: %w\n%s", err, output)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Đăng nhập thành công vào Registry %s", registryServer)))
		fmt.Println("Từ giờ các Node trong cụm đã có thể tự động tải Private Image!")
		return nil
	},
}

func init() {
	registryLoginCmd.Flags().StringVarP(&registryServer, "server", "s", "docker.io", "Địa chỉ Registry (VD: docker.io, ghcr.io)")
	registryLoginCmd.Flags().StringVarP(&registryUser, "user", "u", "", "Tài khoản đăng nhập (Bắt buộc)")
	registryLoginCmd.Flags().StringVarP(&registryPass, "pass", "p", "", "Mật khẩu hoặc Access Token (Bắt buộc)")

	registryLoginCmd.MarkFlagRequired("user")
	registryLoginCmd.MarkFlagRequired("pass")

	registryCmd.AddCommand(registryLoginCmd)
	rootCmd.AddCommand(registryCmd)
}
