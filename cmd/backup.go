package cmd

import (
	"github.com/spf13/cobra"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Quản lý backup & restore",
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Tạo backup toàn bộ cluster data",
	RunE: func(cmd *cobra.Command, args []string) error {
		println(ui.RenderInfo("Backup: tính năng đang được phát triển (v1.1)"))
		return nil
	},
}

var backupRestoreCmd = &cobra.Command{
	Use:   "restore [backup-id]",
	Short: "Restore từ backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		println(ui.RenderInfo("Restore: tính năng đang được phát triển (v1.1)"))
		return nil
	},
}

var backupListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Danh sách các bản backup",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		println(ui.RenderInfo("Backup list: tính năng đang được phát triển (v1.1)"))
		return nil
	},
}

func init() {
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	backupCmd.AddCommand(backupListCmd)
}
