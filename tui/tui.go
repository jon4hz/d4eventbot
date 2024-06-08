package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jon4hz/d4eventbot/core"
)

type (
	ErrMsg    error
	EventsMsg string
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type Model struct {
	viewport viewport.Model
	spinner  spinner.Model
	client   *core.Client
	state    state
	err      error
}

type state int

const (
	stateLoading state = iota + 1
	stateShowEvents
)

func New(client *core.Client) Model {
	m := Model{
		client:   client,
		viewport: viewport.New(0, 0),
		state:    stateLoading,
		spinner:  spinner.New(spinner.WithSpinner(spinner.Meter)),
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Sequence(
		m.spinner.Tick,
		m.FetchEvents(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "ctrl+r":
			m.state = stateLoading
			return m, m.FetchEvents()
		}

	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width - docStyle.GetHorizontalFrameSize()
		m.viewport.Height = msg.Height - docStyle.GetVerticalFrameSize()

	case ErrMsg:
		m.err = msg
		return m, nil

	case EventsMsg:
		m.state = stateShowEvents
		m.viewport.SetContent(string(msg))
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.state {
	case stateLoading:
		return docStyle.Render(m.spinner.View() + "  Loading...")
	case stateShowEvents:
		return docStyle.Render(m.viewport.View())
	}

	return "This should never happen :("
}

func (m Model) FetchEvents() tea.Cmd {
	return func() tea.Msg {
		events, err := m.client.GetMessage()
		if err != nil {
			return ErrMsg(err)
		}
		return EventsMsg(events)
	}
}
