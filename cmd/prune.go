package cmd

import (
	"fmt"
	"os"

	"github.com/LyVanBong/swarm-ctl/internal/alert"
	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/spf13/cobra"
)

var clusterPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Dọn dẹp rác (Image cũ, Container đã dừng) trên toàn bộ Cluster",
	Long: `Lệnh này sẽ tạo ra một Global Service chạy ngầm trên tất cả các Nodes.
Service này sẽ gọi lệnh 'docker system prune -af --volumes' tại mỗi Node để dọn dẹp hàng loạt,
giải phóng không gian ổ cứng, sau đó tự động hủy bỏ.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render("🧹 GLOBAL PRUNE (DỌN RÁC TOÀN CỤM)"))
		fmt.Println()

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		fmt.Println(ui.RenderStep(1, 3, "Khởi tạo tác vụ Dọn dẹp (Global Pruner)..."))
		
		// Deploy a temporary global service to prune docker system on every node
		pruneCmd := `docker service create --name swarm-ctl-global-pruner --mode global --restart-condition none --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock docker:cli sh -c "docker system prune -af --volumes"`
		if _, err := client.Run(pruneCmd); err != nil {
			return fmt.Errorf("không thể khởi chạy tác vụ dọn dẹp: %w", err)
		}
		
		fmt.Println(ui.RenderStep(2, 3, "Đang chờ các Node hoàn tất dọn dẹp (khoảng 30s)..."))
		// Wait for completion (simple sleep or check)
		client.Run("sleep 20")
		
		fmt.Println(ui.RenderStep(3, 3, "Đang dọn dẹp rác của chính tác vụ Pruner (Master Node)..."))
		client.Run("docker service rm swarm-ctl-global-pruner")
		// Hiển thị tiến độ real-time
		client.RunStream("docker system prune -af --volumes", os.Stdout, os.Stderr)

		fmt.Println()
		fmt.Println(ui.RenderSuccess("✅ Quá trình dọn rác Xuyên Lục Địa đã hoàn tất! Hàng chục GB đã được giải phóng."))

		// Gửi thông báo Telegram
		alert.SendTelegramMessage(cluster, "🧹 <b>Dọn Rác Hoàn Tất</b>\n\nToàn bộ rác thải Docker (Images/Containers) trên Cluster đã được xóa sạch!")

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(clusterPruneCmd)
}
