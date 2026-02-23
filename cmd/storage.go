package cmd

import (
	"github.com/spf13/cobra"
	"github.com/softtynet/swarm-ctl/internal/ui"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Quản lý distributed storage (MinIO/GlusterFS)",
}

var storageStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Xem trạng thái storage cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Query MinIO health API
		println(ui.RenderInfo("Storage status: tính năng đang được phát triển (v1.1)"))
		return nil
	},
}

var storageExpandCmd = &cobra.Command{
	Use:   "expand --node [ip]",
	Short: "Thêm storage node mới vào MinIO cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		println(ui.RenderInfo("Storage expand: tính năng đang được phát triển (v1.1)"))
		return nil
	},
}

func init() {
	storageCmd.AddCommand(storageStatusCmd)
	storageCmd.AddCommand(storageExpandCmd)
}
