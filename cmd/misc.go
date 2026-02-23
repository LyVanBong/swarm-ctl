package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
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
	Long:  "Real-time dashboard theo dõi nodes và services",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return fmt.Errorf("không có active cluster — chạy: swarm-ctl cluster init")
		}

		// Fetch function sẽ được gọi mỗi khi refresh
		fetchFn := func() ui.ClusterData {
			client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
			if err := client.Connect(); err != nil {
				return ui.ClusterData{Error: err.Error()}
			}
			defer client.Close()

			// Fetch nodes
			nodeOut, err := client.Run(`docker node ls --format "{{.Hostname}}|{{.Status}}|{{.ManagerStatus}}|{{.Availability}}"`)
			var nodes []ui.NodeRow
			if err == nil {
				for _, line := range strings.Split(strings.TrimSpace(nodeOut), "\n") {
					parts := strings.Split(line, "|")
					if len(parts) >= 4 {
						role := "Worker"
						if parts[2] != "" {
							role = "Manager"
						}
						nodes = append(nodes, ui.NodeRow{
							Hostname:     parts[0],
							Status:       parts[1],
							Role:         role,
							Availability: parts[3],
						})
					}
				}
			}

			// Fetch services
			svcOut, err := client.Run(`docker service ls --format "{{.Name}}|{{.Mode}}|{{.Replicas}}|{{.Image}}"`)
			var services []ui.ServiceRow
			if err == nil {
				for _, line := range strings.Split(strings.TrimSpace(svcOut), "\n") {
					parts := strings.Split(line, "|")
					if len(parts) >= 4 {
						replicas := parts[2]
						repParts := strings.Split(replicas, "/")
						healthy := len(repParts) == 2 && repParts[0] == repParts[1]
						services = append(services, ui.ServiceRow{
							Name:     parts[0],
							Mode:     parts[1],
							Replicas: replicas,
							Image:    parts[3],
							Healthy:  healthy,
						})
					}
				}
			}

			return ui.ClusterData{Nodes: nodes, Services: services}
		}

		model := ui.NewDashboard(cluster.Name, cluster.MasterIP, fetchFn)
		p := tea.NewProgram(model, tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

