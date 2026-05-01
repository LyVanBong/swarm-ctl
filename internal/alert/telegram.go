package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/LyVanBong/swarm-ctl/internal/config"
)

// SendTelegramMessage sends a message to the configured Telegram chat if enabled
func SendTelegramMessage(cluster *config.Cluster, message string) error {
	if cluster == nil || !cluster.Alert.Enabled || cluster.Alert.BotToken == "" || cluster.Alert.ChatID == "" {
		return nil // Alert not configured or disabled
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cluster.Alert.BotToken)

	payload := map[string]interface{}{
		"chat_id":    cluster.Alert.ChatID,
		"text":       fmt.Sprintf("🤖 <b>SWARM-CTL (%s)</b>\n\n%s", cluster.Name, message),
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API trả về mã lỗi: %d", resp.StatusCode)
	}

	return nil
}
