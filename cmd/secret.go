package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/softtynet/swarm-ctl/internal/config"
	"github.com/softtynet/swarm-ctl/internal/ssh"
	"github.com/softtynet/swarm-ctl/internal/ui"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Quản lý Docker Secrets",
}

var secretAddCmd = &cobra.Command{
	Use:   "add [name] [value]",
	Short: "Tạo Docker secret mới",
	Args:  cobra.ExactArgs(2),
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

		secretName, secretValue := args[0], args[1]
		output, err := client.Run(fmt.Sprintf(
			"echo '%s' | docker secret create %s -", secretValue, secretName))
		if err != nil {
			return fmt.Errorf("tạo secret thất bại: %w\n%s", err, output)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Secret '%s' đã được tạo", secretName)))
		return nil
	},
}

var secretListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Xem danh sách secrets",
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

		fmt.Println(ui.SectionHeader.Render(" DOCKER SECRETS "))
		output, err := client.Run(`docker secret ls --format "table {{.Name}}\t{{.CreatedAt}}\t{{.UpdatedAt}}"`)
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil
	},
}

var secretRemoveCmd = &cobra.Command{
	Use:     "remove [name]",
	Short:   "Xóa Docker secret",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
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

		output, err := client.Run(fmt.Sprintf("docker secret rm %s", args[0]))
		if err != nil {
			return fmt.Errorf("xóa secret thất bại: %w\n%s", err, output)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Secret '%s' đã được xóa", args[0])))
		return nil
	},
}

var secretRotateCmd = &cobra.Command{
	Use:   "rotate [name] [new-value]",
	Short: "Rotate secret (tạo version mới, restart services dùng nó)",
	Long: `Rotate secret an toàn:
  1. Tạo secret mới với tên tạm thời
  2. Update service để dùng secret mới
  3. Xóa secret cũ
  
Ví dụ:
  swarm-ctl secret rotate db_password newpassword123`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(ui.RenderInfo("Secret rotation: tính năng đang được phát triển (v1.1)"))
		fmt.Println(ui.RenderInfo("Cho đến khi có: dùng 'docker secret create' thủ công"))
		return nil
	},
}

func init() {
	secretCmd.AddCommand(secretAddCmd)
	secretCmd.AddCommand(secretListCmd)
	secretCmd.AddCommand(secretRemoveCmd)
	secretCmd.AddCommand(secretRotateCmd)
}
