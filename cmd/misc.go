package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Version   = "0.1.0"
	BuildDate = "2026-02-23"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Xem phiên bản hiện tại",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("swarm-ctl v%s (built %s)\n", Version, BuildDate)
		fmt.Println("https://github.com/LyVanBong/swarm-ctl")
	},
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Mở Live TUI Dashboard",
	Long:  "Real-time dashboard theo dõi nodes và services (Bubbletea TUI)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement Bubbletea dashboard (Phase 5)
		fmt.Println("🖥️  Dashboard TUI đang được phát triển...")
		fmt.Println("   Trong khi chờ đợi, dùng:")
		fmt.Println("   → swarm-ctl cluster status")
		fmt.Println("   → swarm-ctl node list")
		fmt.Println("   → swarm-ctl service list")
		return nil
	},
}
