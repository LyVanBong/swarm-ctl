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

// GetCatalog trả về danh sách phong phú các Apps (Top Open Source)
func GetCatalog() []App {
	return []App{
		// --------------------------------------------------------------------------------
		// CATEGORY: TỰ ĐỘNG HÓA & NỘI BỘ (Automation & Internal Tools)
		// --------------------------------------------------------------------------------
		{
			ID:          "n8n",
			Name:        "n8n",
			Category:    "Tự động hóa",
			Description: "Nền tảng Tự động hóa kết nối hơn 300+ APIs phức tạp (Sự thay thế hoàn hảo cho Zapier).",
			Template: `version: "3.8"
services:
  app:
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
			Category:    "Công cụ Nội bộ",
			Description: "Tạo Smart Spreadsheet giống hệt Airtable từ bất kỳ MySQL/Postgres nào chỉ với 1 Click.",
			Template: `version: "3.8"
services:
  app:
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

		// --------------------------------------------------------------------------------
		// CATEGORY: DATABASE & TOOLS (Cơ sở dữ liệu & Công cụ)
		// --------------------------------------------------------------------------------
		{
			ID:          "postgres",
			Name:        "PostgreSQL 16",
			Category:    "Database",
			Description: "Hệ quản trị CSDL quan hệ mạnh mẽ nhất thế giới (Kèm theo pgAdmin).",
			Template: `version: "3.8"
services:
  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=changeme123
      - POSTGRES_DB=appdb
    volumes:
      - postgres_data:/var/lib/postgresql/data
    deploy:
      replicas: 1
      placement:
        constraints: [node.role == manager]
    networks:
      - proxy_public
  pgadmin:
    image: dpage/pgadmin4
    environment:
      - PGADMIN_DEFAULT_EMAIL=admin@{{ .Domain }}
      - PGADMIN_DEFAULT_PASSWORD=changeme123
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.pgadmin.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.pgadmin.entrypoints=https"
        - "traefik.http.routers.pgadmin.tls.certresolver=le"
        - "traefik.http.services.pgadmin.loadbalancer.server.port=80"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public
volumes:
  postgres_data:
networks:
  proxy_public:
    external: true
`,
		},
		{
			ID:          "adminer",
			Name:        "Adminer",
			Category:    "Database Tool",
			Description: "Công cụ quản lý Database (MySQL, Postgres, SQLite) giao diện web siêu nhẹ.",
			Template: `version: "3.8"
services:
  app:
    image: adminer
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.adminer.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.adminer.entrypoints=https"
        - "traefik.http.routers.adminer.tls.certresolver=le"
        - "traefik.http.services.adminer.loadbalancer.server.port=8080"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public
networks:
  proxy_public:
    external: true
`,
		},

		// --------------------------------------------------------------------------------
		// CATEGORY: BACKEND AS A SERVICE (BaaS & API)
		// --------------------------------------------------------------------------------
		{
			ID:          "pocketbase",
			Name:        "PocketBase",
			Category:    "Backend (BaaS)",
			Description: "Backend API siêu nhanh gọn (Auth, Database, Storage) gói gọn trong 1 file duy nhất.",
			Template: `version: "3.8"
services:
  app:
    image: ghcr.io/pocketbase/pocketbase:latest
    command: serve --http=0.0.0.0:8090
    volumes:
      - pocketbase_data:/pb_data
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.pocketbase.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.pocketbase.entrypoints=https"
        - "traefik.http.routers.pocketbase.tls.certresolver=le"
        - "traefik.http.services.pocketbase.loadbalancer.server.port=8090"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public
volumes:
  pocketbase_data:
networks:
  proxy_public:
    external: true
`,
		},
		{
			ID:          "meilisearch",
			Name:        "MeiliSearch",
			Category:    "Search Engine",
			Description: "Search Engine mã nguồn mở, tốc độ phản hồi tính bằng ms, thay thế hoàn hảo cho Elasticsearch.",
			Template: `version: "3.8"
services:
  app:
    image: getmeili/meilisearch:v1.6
    environment:
      - MEILI_MASTER_KEY=SuperSecretKeyChangeme123
      - MEILI_ENV=production
    volumes:
      - meili_data:/meili_data
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.meili.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.meili.entrypoints=https"
        - "traefik.http.routers.meili.tls.certresolver=le"
        - "traefik.http.services.meili.loadbalancer.server.port=7700"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public
volumes:
  meili_data:
networks:
  proxy_public:
    external: true
`,
		},

		// --------------------------------------------------------------------------------
		// CATEGORY: PHÂN TÍCH & GIÁM SÁT (Analytics & Monitoring)
		// --------------------------------------------------------------------------------
		{
			ID:          "metabase",
			Name:        "Metabase",
			Category:    "Analytic (Dữ liệu)",
			Description: "Business Intelligence, vẽ BI Chart, Report tự động từ SQL Database.",
			Template: `version: "3.8"
services:
  app:
    image: metabase/metabase:latest
    environment:
      - MB_SITE_URL=https://{{ .Domain }}
    volumes:
      - metabase_data:/metabase-data
    deploy:
      resources:
        limits: { memory: 2048M }
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
		{
			ID:          "uptime-kuma",
			Name:        "Uptime Kuma",
			Category:    "Giám sát",
			Description: "Dashboard Self-hosted tuyệt đẹp theo dõi Uptime, ping, HTTP(s) tương tự (Pingdom).",
			Template: `version: "3.8"
services:
  app:
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
			ID:          "umami",
			Name:        "Umami Analytics",
			Category:    "Analytic (Web)",
			Description: "Công cụ theo dõi lượng truy cập Website cực kì riêng tư, bảo mật, nhẹ nhàng để thay thế Google Analytics.",
			Template: `version: "3.8"
services:
  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=umami
      - POSTGRES_USER=umami
      - POSTGRES_PASSWORD=umami_secret
    volumes:
      - umami_db_data:/var/lib/postgresql/data
    networks:
      - proxy_public
    deploy:
      placement:
        constraints: [node.role == manager]

  app:
    image: ghcr.io/umami-software/umami:postgresql-latest
    environment:
      - DATABASE_URL=postgresql://umami:umami_secret@db:5432/umami
      - DATABASE_TYPE=postgresql
      - APP_SECRET=replace_me_with_a_random_string_later
    depends_on:
      - db
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.umami.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.umami.entrypoints=https"
        - "traefik.http.routers.umami.tls.certresolver=le"
        - "traefik.http.services.umami.loadbalancer.server.port=3000"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public

volumes:
  umami_db_data:
networks:
  proxy_public:
    external: true
`,
		},

		// --------------------------------------------------------------------------------
		// CATEGORY: LƯU TRỮ CHIA SẺ & BẢO MẬT (Storage / Security)
		// --------------------------------------------------------------------------------
		{
			ID:          "nextcloud",
			Name:        "Nextcloud AIO",
			Category:    "Cloud Storage",
			Description: "Google Drive phiên bản tự cài đặt. Quản lý File, Hình ảnh, Danh bạ của riêng bạn.",
			Template: `version: "3.8"
services:
  app:
    image: nextcloud:apache
    environment:
      - NEXTCLOUD_TRUSTED_DOMAINS={{ .Domain }}
    volumes:
      - nextcloud_data:/var/www/html
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.nextcloud.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.nextcloud.entrypoints=https"
        - "traefik.http.routers.nextcloud.tls.certresolver=le"
        - "traefik.http.services.nextcloud.loadbalancer.server.port=80"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public
volumes:
  nextcloud_data:
networks:
  proxy_public:
    external: true
`,
		},
		{
			ID:          "vaultwarden",
			Name:        "Vaultwarden",
			Category:    "Bảo mật",
			Description: "Trình quản lý Mật khẩu tương thích hoàn toàn với ứng dụng Bitwarden.",
			Template: `version: "3.8"
services:
  app:
    image: vaultwarden/server:latest
    environment:
      - DOMAIN=https://{{ .Domain }}
      - SIGNUPS_ALLOWED=true
    volumes:
      - vw_data:/data/
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.vaultwarden.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.vaultwarden.entrypoints=https"
        - "traefik.http.routers.vaultwarden.tls.certresolver=le"
        - "traefik.http.services.vaultwarden.loadbalancer.server.port=80"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public
volumes:
  vw_data:
networks:
  proxy_public:
    external: true
`,
		},

		// --------------------------------------------------------------------------------
		// CATEGORY: WEBSITE & DEV TOOLS (CMS & CI/CD)
		// --------------------------------------------------------------------------------
		{
			ID:          "wordpress",
			Name:        "WordPress",
			Category:    "CMS (Website)",
			Description: "Nền tảng tạo Blog/Website phổ biến nhất thế giới (Bao gồm MariaDB).",
			Template: `version: "3.8"
services:
  db:
    image: mariadb:10.6
    environment:
      - MYSQL_ROOT_PASSWORD=somewordpress
      - MYSQL_DATABASE=wordpress
      - MYSQL_USER=wordpress
      - MYSQL_PASSWORD=wordpress
    volumes:
      - wp_db_data:/var/lib/mysql
    networks:
      - proxy_public
    deploy:
      placement:
        constraints: [node.role == manager]

  app:
    image: wordpress:latest
    environment:
      - WORDPRESS_DB_HOST=db
      - WORDPRESS_DB_USER=wordpress
      - WORDPRESS_DB_PASSWORD=wordpress
      - WORDPRESS_DB_NAME=wordpress
    volumes:
      - wp_data:/var/www/html
    depends_on:
      - db
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.wp.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.wp.entrypoints=https"
        - "traefik.http.routers.wp.tls.certresolver=le"
        - "traefik.http.services.wp.loadbalancer.server.port=80"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public

volumes:
  wp_db_data:
  wp_data:
networks:
  proxy_public:
    external: true
`,
		},
		{
			ID:          "ghost",
			Name:        "Ghost CMS",
			Category:    "CMS (Website)",
			Description: "Nền tảng xuất bản Nội dung, viết Blog siêu tốc độ và hiện đại (Nhẹ hơn WP).",
			Template: `version: "3.8"
services:
  app:
    image: ghost:latest
    environment:
      - url=https://{{ .Domain }}
    volumes:
      - ghost_data:/var/lib/ghost/content
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.ghost.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.ghost.entrypoints=https"
        - "traefik.http.routers.ghost.tls.certresolver=le"
        - "traefik.http.services.ghost.loadbalancer.server.port=2368"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public
volumes:
  ghost_data:
networks:
  proxy_public:
    external: true
`,
		},
		{
			ID:          "gitea",
			Name:        "Gitea",
			Category:    "Dev Tools (Git)",
			Description: "Máy chủ lưu trữ Code Git cục bộ (Self-hosted Github/Gitlab) cực kì nhẹ.",
			Template: `version: "3.8"
services:
  app:
    image: gitea/gitea:latest
    environment:
      - USER_UID=1000
      - USER_GID=1000
      - GITEA__server__DOMAIN={{ .Domain }}
      - GITEA__server__ROOT_URL=https://{{ .Domain }}/
    volumes:
      - gitea_data:/data
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.gitea.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.gitea.entrypoints=https"
        - "traefik.http.routers.gitea.tls.certresolver=le"
        - "traefik.http.services.gitea.loadbalancer.server.port=3000"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public
volumes:
  gitea_data:
networks:
  proxy_public:
    external: true
`,
		},
		{
			ID:          "rabbitmq",
			Name:        "RabbitMQ",
			Category:    "Message Queue",
			Description: "Broker điều phối Message phổ biến nhất (Có sẵn trang quản lý Management UI).",
			Template: `version: "3.8"
services:
  app:
    image: rabbitmq:3-management-alpine
    environment:
      - RABBITMQ_DEFAULT_USER=admin
      - RABBITMQ_DEFAULT_PASS=admin_pass_123
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    deploy:
      replicas: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.rabbitmq.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.rabbitmq.entrypoints=https"
        - "traefik.http.routers.rabbitmq.tls.certresolver=le"
        - "traefik.http.services.rabbitmq.loadbalancer.server.port=15672"
        - "traefik.swarm.network=proxy_public"
    networks:
      - proxy_public
volumes:
  rabbitmq_data:
networks:
  proxy_public:
    external: true
`,
		},
	}
}

// GenerateYaml sinh ra file Compose string dựa theo config Domain do người dùng truyền vào
func GenerateYaml(id string, config AppConfig) (string, error) {
	catalogList := GetCatalog()
	var selectedApp *App

	for _, app := range catalogList {
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
