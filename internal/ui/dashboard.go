package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"carbonqt/internal/energy"
	"carbonqt/internal/models"
	"carbonqt/internal/monitor"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DashboardConfig struct {
	Estimator energy.Estimator
	RepoRoot  string
	Refresh   time.Duration
}

type dashboardModel struct {
	config      DashboardConfig
	system      models.SystemMetrics
	processes   []models.ProcessMetrics
	totalCarbon float64
	trend       []float64
	width       int
	height      int
	selected    int
	status      string
	err         error
}

type dashboardTickMsg time.Time

func StartDashboard(config DashboardConfig) error {
	model := dashboardModel{config: config}
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

func (m dashboardModel) Init() tea.Cmd {
	return tea.Tick(m.config.Refresh, func(t time.Time) tea.Msg {
		return dashboardTickMsg(t)
	})
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dashboardTickMsg:
		system, err := monitor.GetSystemMetrics()
		if err != nil {
			m.err = err
			return m, nil
		}

		processes, err := monitor.ListProcesses(m.config.RepoRoot)
		if err != nil {
			m.err = err
			return m, nil
		}

		total, processes := m.config.Estimator.ApplyCarbon(processes, m.config.Refresh)
		sort.Slice(processes, func(i, j int) bool { return processes[i].CarbonKg > processes[j].CarbonKg })
		if len(processes) > 10 {
			processes = processes[:10]
		}

		if m.selected >= len(processes) {
			m.selected = len(processes) - 1
		}
		if m.selected < 0 {
			m.selected = 0
		}

		maxSamples := int((30 * time.Second) / m.config.Refresh)
		if maxSamples < 1 {
			maxSamples = 1
		}
		m.trend = append(m.trend, total)
		if len(m.trend) > maxSamples {
			m.trend = m.trend[len(m.trend)-maxSamples:]
		}

		m.system = system
		m.totalCarbon = total
		m.processes = processes
		return m, tea.Tick(m.config.Refresh, func(t time.Time) tea.Msg {
			return dashboardTickMsg(t)
		})
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.selected > 0 {
				m.selected--
			}
		case "down":
			if m.selected < len(m.processes)-1 {
				m.selected++
			}
		case "k":
			if len(m.processes) == 0 || m.selected < 0 || m.selected >= len(m.processes) {
				m.status = "No process selected."
				return m, nil
			}
			pid := m.processes[m.selected].PID
			name := m.processes[m.selected].Name
			if err := monitor.KillProcess(pid); err != nil {
				m.status = fmt.Sprintf("Failed to kill %s (PID %d).", name, pid)
				return m, nil
			}
			m.status = fmt.Sprintf("Killed %s (PID %d).", name, pid)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m dashboardModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	systemInfoBlock := strings.Join([]string{
		fmt.Sprintf("CPU: %s (%d cores)", fallback(m.system.CPUModel, "Unknown"), m.system.CPUCores),
		fmt.Sprintf("RAM: %s total", formatBytes(m.system.MemoryTotalBytes)),
		fmt.Sprintf("Platform: %s", fallback(m.system.Platform, "Unknown")),
		fmt.Sprintf("Uptime: %s", formatDuration(time.Duration(m.system.UptimeSeconds)*time.Second)),
	}, "\n")

	systemBlock := strings.Join([]string{
		"CPU Usage " + progressBar(24, m.system.CPUPercent) + fmt.Sprintf(" %.1f%%", m.system.CPUPercent),
		"RAM Usage " + progressBar(24, m.system.MemoryPercent) + fmt.Sprintf(" %.1f%%", m.system.MemoryPercent),
	}, "\n")

	carbonBlock := fmt.Sprintf("Estimated Emissions: %s mg", formatCarbonMg(m.totalCarbon))

	var topProcessBlock string
	if len(m.processes) > 0 {
		top := m.processes[0]
		topProcessBlock = strings.Join([]string{
			"Highest Carbon Process",
			fallback(top.Name, "Unknown"),
			fmt.Sprintf("Power: %.2f W", top.PowerW),
			fmt.Sprintf("Carbon: %s mg", formatCarbonMg(top.CarbonKg)),
		}, "\n")
	} else {
		topProcessBlock = "Highest Carbon Process\n(n/a)"
	}

	processBlock := RenderDashboardTableSelected(m.processes, m.selected, 10)

	width := m.width
	if width <= 0 {
		width = 120
	}
	leftWidth := (width - 6) / 2
	if leftWidth < 50 {
		leftWidth = 50
	}
	panelBox := lipgloss.NewStyle().Width(leftWidth)
	systemPanel := panelBox.Render(headerStyle.Render("System") + "\n" + systemInfoBlock + "\n" + systemBlock)
	carbonPanel := panelBox.Render(headerStyle.Render("Estimated Carbon Emissions") + "\n" + carbonBlock + "\n\n" + headerStyle.Render("Top Carbon Process") + "\n" + topProcessBlock)

	upperRow := lipgloss.JoinHorizontal(lipgloss.Top, systemPanel, carbonPanel)

	ribbon := "Up ^   Down v   K Kill   Q Quit"
	if strings.TrimSpace(m.status) != "" {
		ribbon = m.status + "  |  " + ribbon
	}
	ribbonStyle := lipgloss.NewStyle().Background(lipgloss.Color("231")).Foreground(lipgloss.Color("16")).Bold(true).Padding(0, 1).Align(lipgloss.Center)
	ribbon = ribbonStyle.Width(width).Render(ribbon)

	title := titleStyle.Width(width).Align(lipgloss.Center).Render("CarbonQt Dashboard")
	sections := []string{
		title,
		upperRow,
		headerStyle.Render("Processes"),
		panelStyle.Render(processBlock),
		ribbon,
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(sections, "\n\n"))
}
