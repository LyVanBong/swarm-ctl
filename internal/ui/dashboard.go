package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─────────────────────────────────────────────────────────────
// Messages
// ─────────────────────────────────────────────────────────────

type ClusterData struct {
	Nodes    []NodeRow
	Services []ServiceRow
	Error    string
}

type NodeRow struct {
	Hostname   string
	Status     string
	Role       string
	CPU        string
	Memory     string
	Availability string
}

type ServiceRow struct {
	Name     string
	Mode     string
	Replicas string
	Image    string
	Healthy  bool
}

type refreshMsg ClusterData
type tickMsg struct{}

// ─────────────────────────────────────────────────────────────
// Dashboard Model
// ─────────────────────────────────────────────────────────────

type DashboardModel struct {
	clusterName string
	masterIP    string
	fetchFn     func() ClusterData

	nodes    []NodeRow
	services []ServiceRow
	lastErr  string

	nodeTable    table.Model
	serviceTable table.Model
	spinner      spinner.Model

	loading    bool
	activeTab  int // 0=nodes, 1=services
	width      int
	height     int
	quitting   bool
}

func NewDashboard(clusterName, masterIP string, fetchFn func() ClusterData) DashboardModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	nodeT := table.New(
		table.WithColumns([]table.Column{
			{Title: "HOSTNAME", Width: 25},
			{Title: "STATUS", Width: 10},
			{Title: "ROLE", Width: 10},
			{Title: "AVAILABILITY", Width: 14},
		}),
		table.WithFocused(true),
		table.WithHeight(10),
		table.WithStyles(tableStyles()),
	)

	svcT := table.New(
		table.WithColumns([]table.Column{
			{Title: "SERVICE", Width: 30},
			{Title: "MODE", Width: 12},
			{Title: "REPLICAS", Width: 12},
			{Title: "IMAGE", Width: 40},
		}),
		table.WithFocused(false),
		table.WithHeight(15),
		table.WithStyles(tableStyles()),
	)

	return DashboardModel{
		clusterName:  clusterName,
		masterIP:     masterIP,
		fetchFn:      fetchFn,
		nodeTable:    nodeT,
		serviceTable: svcT,
		spinner:      s,
		loading:      true,
		activeTab:    0,
	}
}

func tableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorPrimary).
		BorderBottom(true).
		Bold(true).
		Foreground(ColorPrimary)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#ffffff")).
		Background(ColorPrimary).
		Bold(true)
	return s
}

// ─────────────────────────────────────────────────────────────
// Init, Update, View
// ─────────────────────────────────────────────────────────────

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchData(),
	)
}

func (m DashboardModel) fetchData() tea.Cmd {
	return func() tea.Msg {
		data := m.fetchFn()
		return refreshMsg(data)
	}
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "tab", "1", "2":
			if msg.String() == "1" {
				m.activeTab = 0
			} else if msg.String() == "2" {
				m.activeTab = 1
			} else {
				m.activeTab = (m.activeTab + 1) % 2
			}
			m.nodeTable.Blur()
			m.serviceTable.Blur()
			if m.activeTab == 0 {
				m.nodeTable.Focus()
			} else {
				m.serviceTable.Focus()
			}

		case "r":
			m.loading = true
			cmds = append(cmds, m.fetchData())
		}

	case refreshMsg:
		m.loading = false
		m.lastErr = msg.Error
		m.nodes = msg.Nodes
		m.services = msg.Services

		// Update node table
		nodeRows := make([]table.Row, len(m.nodes))
		for i, n := range m.nodes {
			nodeRows[i] = table.Row{
				n.Hostname,
				n.Status,
				n.Role,
				n.Availability,
			}
		}
		m.nodeTable.SetRows(nodeRows)

		// Update service table
		svcRows := make([]table.Row, len(m.services))
		for i, s := range m.services {
			svcRows[i] = table.Row{
				s.Name,
				s.Mode,
				s.Replicas,
				truncate(s.Image, 38),
			}
		}
		m.serviceTable.SetRows(svcRows)


	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Forward keystrokes to active table
	var tableCmd tea.Cmd
	if m.activeTab == 0 {
		m.nodeTable, tableCmd = m.nodeTable.Update(msg)
	} else {
		m.serviceTable, tableCmd = m.serviceTable.Update(msg)
	}
	cmds = append(cmds, tableCmd)

	return m, tea.Batch(cmds...)
}

func (m DashboardModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// ── Header ──────────────────────────────────────────────
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(ColorPrimary).
		Padding(0, 2).
		Width(m.width).
		Render(fmt.Sprintf(
			"🐳 SWARM-CTL  │  Cluster: %s  │  Master: %s",
			m.clusterName, m.masterIP,
		))
	b.WriteString(header + "\n")

	if m.lastErr != "" {
		b.WriteString(Danger.Render("❌ Error: "+m.lastErr) + "\n")
	}

	// ── Tabs ─────────────────────────────────────────────────
	tabs := m.renderTabs()
	b.WriteString(tabs + "\n\n")

	if m.loading {
		b.WriteString(fmt.Sprintf("  %s Đang tải dữ liệu...\n", m.spinner.View()))
		return b.String()
	}

	// ── Content ───────────────────────────────────────────────
	if m.activeTab == 0 {
		// Nodes tab
		b.WriteString(lipgloss.NewStyle().MarginLeft(1).Render(m.nodeTable.View()))
		b.WriteString("\n\n")
		// Summary
		healthy := 0
		for _, n := range m.nodes {
			if n.Status == "Ready" {
				healthy++
			}
		}
		b.WriteString(fmt.Sprintf("  Nodes: %s healthy / %s total\n",
			Success.Render(fmt.Sprintf("%d", healthy)),
			Bold.Render(fmt.Sprintf("%d", len(m.nodes)))))
	} else {
		// Services tab
		b.WriteString(lipgloss.NewStyle().MarginLeft(1).Render(m.serviceTable.View()))
		b.WriteString("\n\n")
		// Summary
		healthy := 0
		for _, s := range m.services {
			if s.Healthy {
				healthy++
			}
		}
		b.WriteString(fmt.Sprintf("  Services: %s healthy / %s total\n",
			Success.Render(fmt.Sprintf("%d", healthy)),
			Bold.Render(fmt.Sprintf("%d", len(m.services)))))
	}

	// ── Footer / Help ─────────────────────────────────────────
	footer := Muted.Render("  [Tab] Switch  [1] Nodes  [2] Services  [r] Refresh  [q] Quit")
	b.WriteString("\n" + footer)

	return b.String()
}

func (m DashboardModel) renderTabs() string {
	tabs := []string{"1. Nodes", "2. Services"}
	var rendered []string
	for i, t := range tabs {
		if i == m.activeTab {
			rendered = append(rendered, lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(ColorPrimary).
				Padding(0, 2).
				Render(t))
		} else {
			rendered = append(rendered, lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 2).
				Border(lipgloss.NormalBorder(), false, false, false, false).
				Render(t))
		}
	}
	return "  " + strings.Join(rendered, "  ")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
