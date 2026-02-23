package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/spf13/cobra"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Cập nhật swarm-ctl lên phiên bản mới nhất từ Github",
	Long:  "Lệnh này sẽ tải về bản cập nhật mới nhất (Latest Release) của công cụ từ Github và tự động thay thế phiên bản cũ.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(ui.RenderStep(1, 4, "Kiểm tra phiên bản mới nhất..."))

		// Lấy latest release
		resp, err := http.Get("https://api.github.com/repos/LyVanBong/swarm-ctl/releases/latest")
		if err != nil {
			return fmt.Errorf("không thể kết nối Github API: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Github API trả về lỗi: %v", resp.Status)
		}

		var release githubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return fmt.Errorf("lỗi đọc dữ liệu Github: %v", err)
		}

		latestVersion := release.TagName
		currentVersion := Version

		if currentVersion == latestVersion || "v"+currentVersion == latestVersion {
			fmt.Println(ui.RenderSuccess(fmt.Sprintf("Bạn đang ở phiên bản mới nhất: %s 🎉", currentVersion)))
			return nil
		}

		fmt.Printf("🚀 Phát hiện phiên bản mới: %s (Hiện tại: %s)\n", ui.Info.Render(latestVersion), ui.Warning.Render(currentVersion))
		fmt.Println(ui.RenderStep(2, 4, "Tải tệp tin cập nhật từ Github..."))

		osName := runtime.GOOS
		archName := runtime.GOARCH
		fileName := fmt.Sprintf("swarm-ctl-%s-%s", osName, archName)
		downloadURL := fmt.Sprintf("https://github.com/LyVanBong/swarm-ctl/releases/latest/download/%s", fileName)

		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("không thể định vị thư mục thực thi: %v", err)
		}
		exePath, err = filepath.EvalSymlinks(exePath)
		if err != nil {
			return err
		}

		// Download to temp file
		tmpFile := exePath + ".tmp"
		out, err := os.Create(tmpFile)
		if err != nil {
			return fmt.Errorf("không có quyền ghi vào hệ thống. Gợi ý: Hãy thêm 'sudo' trước câu lệnh: sudo swarm-ctl update")
		}
		defer out.Close()

		dlResp, err := http.Get(downloadURL)
		if err != nil {
			return fmt.Errorf("lỗi tải xuống bản cập nhật: %v", err)
		}
		defer dlResp.Body.Close()

		if dlResp.StatusCode != http.StatusOK {
			return fmt.Errorf("không tìm thấy file chạy '%s' cho hệ điều hành của bạn. Mã lỗi: %d", fileName, dlResp.StatusCode)
		}

		_, err = io.Copy(out, dlResp.Body)
		if err != nil {
			return fmt.Errorf("lỗi ghi file cập nhật: %v", err)
		}
		out.Close() // Close stream trước để sửa quyền

		fmt.Println(ui.RenderStep(3, 4, "Thay thế mã nguồn và cấp quyền..."))

		if err := os.Chmod(tmpFile, 0755); err != nil {
			return fmt.Errorf("lỗi phân quyền thực thi: %v", err)
		}

		if err := os.Rename(tmpFile, exePath); err != nil {
			return fmt.Errorf("không thể đè tệp cập nhật lên file hiện tại. Gợi ý: Dùng lệnh 'sudo swarm-ctl update'. Lỗi chi tiết: %v", err)
		}

		fmt.Println(ui.RenderStep(4, 4, "Hoàn tất giải nén!"))
		fmt.Println()
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Nâng cấp thành công lên phiên bản %s", latestVersion)))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
