package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// ServiceSpec là định nghĩa của 1 service
type ServiceSpec struct {
	Name          string
	Image         string
	Domain        string
	Port          int
	Replicas      int
	CPULimit      string
	MemoryLimit   string
	CPUReserve    string
	MemReserve    string
	Placement     string // "tier=app" hoặc "node.role==manager"
	Middlewares   []string
	Env           map[string]string
	Secrets       []string
	Volumes       []string
	Networks      []string
	RestartPolicy string
	UpdateOrder   string
}

// GenerateCompose tạo nội dung docker-compose.yml từ ServiceSpec
func GenerateCompose(spec ServiceSpec) (string, error) {
	// Set defaults
	if spec.Replicas == 0 {
		spec.Replicas = 1
	}
	if spec.CPULimit == "" {
		spec.CPULimit = "0.5"
	}
	if spec.MemoryLimit == "" {
		spec.MemoryLimit = "512M"
	}
	if spec.CPUReserve == "" {
		spec.CPUReserve = "0.1"
	}
	if spec.MemReserve == "" {
		spec.MemReserve = "128M"
	}
	if spec.RestartPolicy == "" {
		spec.RestartPolicy = "on-failure"
	}
	if spec.UpdateOrder == "" {
		spec.UpdateOrder = "start-first"
	}

	tmpl := template.Must(template.New("compose").Funcs(template.FuncMap{
		"indent": func(spaces int, s string) string {
			pad := strings.Repeat(" ", spaces)
			return pad + strings.ReplaceAll(s, "\n", "\n"+pad)
		},
		"join": strings.Join,
	}).Parse(composeTemplate))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, spec); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	return buf.String(), nil
}

const composeTemplate = `version: '3.8'

services:
  {{ .Name }}:
    image: {{ .Image }}
{{ if .Env }}    environment:
{{ range $key, $val := .Env }}      - {{ $key }}={{ $val }}
{{ end }}{{ end }}{{ if .Secrets }}    secrets:
{{ range .Secrets }}      - {{ . }}
{{ end }}{{ end }}{{ if .Volumes }}    volumes:
{{ range .Volumes }}      - {{ . }}
{{ end }}{{ end }}    networks:
      - proxy_public
{{ range .Networks }}      - {{ . }}
{{ end }}    deploy:
      replicas: {{ .Replicas }}
      restart_policy:
        condition: {{ .RestartPolicy }}
        delay: 5s
        max_attempts: 3
        window: 120s
      update_config:
        parallelism: 1
        delay: 10s
        order: {{ .UpdateOrder }}
        failure_action: rollback
      resources:
        limits:
          cpus: '{{ .CPULimit }}'
          memory: {{ .MemoryLimit }}
        reservations:
          cpus: '{{ .CPUReserve }}'
          memory: {{ .MemReserve }}
{{ if .Placement }}      placement:
        constraints:
          - node.labels.{{ .Placement }}
{{ end }}      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.{{ .Name }}.rule=Host(` + "`" + `{{ .Domain }}` + "`" + `)"
        - "traefik.http.routers.{{ .Name }}.entrypoints=https"
        - "traefik.http.routers.{{ .Name }}.tls=true"
        - "traefik.http.routers.{{ .Name }}.tls.certresolver=le"
{{ if .Middlewares }}        - "traefik.http.routers.{{ .Name }}.middlewares={{ join .Middlewares "," }}"
{{ end }}        - "traefik.http.services.{{ .Name }}.loadbalancer.server.port={{ .Port }}"
        - "traefik.swarm.network=proxy_public"

networks:
  proxy_public:
    external: true
{{ range .Networks }}  {{ . }}:
    driver: overlay
{{ end }}{{ if .Secrets }}
secrets:
{{ range .Secrets }}  {{ . }}:
    external: true
{{ end }}{{ end }}`
