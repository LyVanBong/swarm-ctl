package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Xem nhật ký kiểm toán (Audit Logs)",
	Long: `In ra lịch sử hành động ghi nhận từ tất cả các lần gọi CLI.
Mỗi lệnh nguy hiểm đều được truy vết tự động.

Ví dụ:
  swarm-ctl audit
  swarm-ctl audit --tail 50`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tailFlag, _ := cmd.Flags().GetInt("tail")

		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		auditFile := filepath.Join(home, ".swarm-ctl", "audit.log")

		data, err := ioutil.ReadFile(auditFile)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(ui.RenderInfo("Log kiểm toán trống (chưa có lệnh nào được lưu)"))
				return nil
			}
			return fmt.Errorf("không thể đọc file audit: %s", err)
		}

		fmt.Println(ui.SectionHeader.Render(" LỊCH SỬ KIỂM TOÁN (AUDIT LOGS) "))

		content := string(data)
		lines := strings.Split(strings.TrimSpace(content), "\n")
		
		if len(lines) == 0 {
			fmt.Println("No logs found.")
			return nil
		}

		// Trimming
		start := 0
		if tailFlag > 0 && len(lines) > tailFlag {
			start = len(lines) - tailFlag
		}

		for i := start; i < len(lines); i++ {
			if lines[i] == "" {
				continue
			}
			fmt.Println("  " + ui.Muted.Render(lines[i]))
		}

		return nil
	},
}

func init() {
	auditCmd.Flags().IntP("tail", "n", 30, "Số lượng dòng log hiển thị")
	rootCmd.AddCommand(auditCmd)
}
