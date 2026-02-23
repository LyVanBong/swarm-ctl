package catalog

import (
	"bytes"
	"fmt"
	"text/template"
)

// App mô tả một ứng dụng có sẵn trong thư viện Store
type App struct {
	ID          string
	Name        string
	Category    string
	Description string
	Template    string
}

// Config cấu hình nhập khi cài đặt
type AppConfig struct {
	Domain string
}

// GetCatalog trả về danh sách các Apps hỗ trợ cài đặt 1 chạm
func GetCatalog() []App {
	return []App{
		{
			ID:          "n8n",
			Name:        "n8n (Workflow Automation)",
			Category:    "Tự động hóa",
			Description: "Nền tảng Tự động hóa kết nối hơn 300+ APIs phức tạp (Sự thay thế hoàn hảo cho Zapier).",
			Template: `version: "3.8"
services:
  n8n:
    image: docker.n8n.io/n8nio/n8n
    command: /bin/sh -c "n8n start"
    environment:
      - N8N_HOST={{ .Domain }}
      - N8N_PORT=5678
      - N8N_PROTOCOL=https
      - NODE_ENV=production
      - WEBHOOK_URL=https://{{ .Domain }}/
      - GENERIC_TIMEZONE=Asia/Ho_Chi_Minh
    volumes:
      - n8n_data:/home/node/.n8n
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.n8n.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.n8n.entrypoints=https"
        - "traefik.http.routers.n8n.tls.certresolver=le"
        - "traefik.http.services.n8n.loadbalancer.server.port=5678"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public

volumes:
  n8n_data:

networks:
  proxy_public:
    external: true
`,
		},
		{
			ID:          "nocodb",
			Name:        "NocoDB",
			Category:    "Cơ sở dữ liệu",
			Description: "Tạo Smart Spreadsheet giống hệt Airtable từ bất kỳ MySQL/Postgres nào chỉ với 1 Click.",
			Template: `version: "3.8"
services:
  nocodb:
    image: nocodb/nocodb:latest
    environment:
      - NC_PUBLIC_URL=https://{{ .Domain }}
      - NC_MIN_THREADS=2
    volumes:
      - nocodb_data:/usr/app/data
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.nocodb.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.nocodb.entrypoints=https"
        - "traefik.http.routers.nocodb.tls.certresolver=le"
        - "traefik.http.services.nocodb.loadbalancer.server.port=8080"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public

volumes:
  nocodb_data:

networks:
  proxy_public:
    external: true
`,
		},
		{
			ID:          "uptime-kuma",
			Name:        "Uptime Kuma",
			Category:    "Giám sát",
			Description: "Dashboard Self-hosted tuyệt đẹp theo dõi Uptime, ping, HTTP(s) tương tự (Pingdom/StatusCake).",
			Template: `version: "3.8"
services:
  uptime-kuma:
    image: louislam/uptime-kuma:1
    volumes:
      - uptime_kuma_data:/app/data
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.uptimekuma.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.uptimekuma.entrypoints=https"
        - "traefik.http.routers.uptimekuma.tls.certresolver=le"
        - "traefik.http.services.uptimekuma.loadbalancer.server.port=3001"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public

volumes:
  uptime_kuma_data:

networks:
  proxy_public:
    external: true
`,
		},
		{
			ID:          "metabase",
			Name:        "Metabase",
			Category:    "Analytic (Dữ liệu)",
			Description: "Business Intelligence, vẽ BI Chart, Report tự động từ SQL Database (Mạnh mẽ ngang Tableau/PowerBI).",
			Template: `version: "3.8"
services:
  metabase:
    image: metabase/metabase:latest
    environment:
      - MB_SITE_URL=https://{{ .Domain }}
    volumes:
      - metabase_data:/metabase-data
    deploy:
      # Metabase cần nhiều RAM (Tối thiểu 1-2GB)
      resources:
        limits:
          memory: 2048M
        reservations:
          memory: 1024M
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.metabase.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.metabase.entrypoints=https"
        - "traefik.http.routers.metabase.tls.certresolver=le"
        - "traefik.http.services.metabase.loadbalancer.server.port=3000"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public

volumes:
  metabase_data:

networks:
  proxy_public:
    external: true
`,
		},
	}
}

// GenerateYaml sinh ra file Compose string dựa theo config Domain do người dùng truyền vào
func GenerateYaml(id string, config AppConfig) (string, error) {
	catalog := GetCatalog()
	var selectedApp *App

	for _, app := range catalog {
		if app.ID == id {
			selectedApp = &app
			break
		}
	}

	if selectedApp == nil {
		return "", fmt.Errorf("không tìm thấy app có ID: %s. Dùng lệnh `swarm-ctl app list` để xem.", id)
	}

	tmpl, err := template.New("compose").Parse(selectedApp.Template)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return "", err
	}

	return buf.String(), nil
}
