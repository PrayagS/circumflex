package bubble

import (
	"clx/bubble/list"
	"clx/cli"
	"clx/comment"
	"clx/core"
	"clx/hn/services/cheeaun"
	"clx/screen"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
)

var docStyle = lipgloss.NewStyle()

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

type editorFinishedMsg struct{ err error }
type enteringCommentSectionMsg struct{ id int }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.list.OnStartup() {
		var cmds []tea.Cmd

		m.list.SetSize(screen.GetTerminalWidth(), screen.GetTerminalHeight())

		spinnerCmd := m.list.StartSpinner()
		cmds = append(cmds, spinnerCmd)

		m.list.SetOnStartup(false)

		fetchCmd := m.list.FetchFrontPageStories()
		cmds = append(cmds, fetchCmd)

		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case enteringCommentSectionMsg:
		cmd := openEditor(msg.id)

		return m, cmd
	case editorFinishedMsg:
		m.list.SetIsVisible(true)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func openEditor(id int) tea.Cmd {
	comments := new(cheeaun.Service).FetchStory(id)

	screenWidth := screen.GetTerminalWidth()
	commentTree := comment.ToString(comments, core.GetConfigWithDefaults(), screenWidth, 0)

	c := cli.WrapLess(commentTree)

	return tea.Exec(tea.WrapExecCommand(c), func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func Run(config *core.Config) {
	cli.ClearScreen()

	m := model{list: list.New(list.NewDefaultDelegate(), config, 0, 0)}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}
