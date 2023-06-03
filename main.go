package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/gj1118/automata/appmodels"
	"github.com/gj1118/automata/styles"
)

type sessionState int

type model struct {
	list         list.Model
	keys         *listKeyMap
	delegateKeys *delegateKeyMap
	keyMap       *appmodels.KeyMap
	styles       styles.Styles
}

type item struct {
	title       string
	description string
}

type delegateKeyMap struct {
	choose key.Binding
	remove key.Binding
}

type listKeyMap struct {
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
	launchItem       key.Binding
}

const (
	mainModel sessionState = iota
	timerModel
	aboutModel
	simpleListModel
)

var (
	current = mainModel
	models  []tea.Model
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render

	color     = termenv.EnvColorProfile().Color
	keyword   = termenv.Style{}.Foreground(color("204")).Background(color("235")).Styled
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var localItem item
		if i, ok := m.SelectedItem().(item); ok {
			localItem = i
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.choose):
				return m.NewStatusMessage(statusMessageStyle("You chose ", localItem.title))
			}
		}

		return nil
	}

	help := []key.Binding{keys.choose, keys.remove}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
		d.remove,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
			d.remove,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter (a)", "choose"),
		),
	}
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

// type menuModel models.MenuModel

func MainModel() (tea.Model, tea.Cmd) {
	initMenu()
	current = mainModel
	return models[current], models[current].Init()
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		toggleSpinner: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle spinner"),
		),
		toggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		toggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		togglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
		launchItem: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "launch item")),
	}
}

func HelpMenu(view ...string) string {
	if len(view) != 0 {
		return helpStyle(fmt.Sprintf("right/l: next • left/h: previous • enter: new %s", view[0]))
	}
	return helpStyle("right/l: next • left/h: previous")

}

func initMenu() []tea.Model {
	var (
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
	)

	// Make initial list of items
	items := []list.Item{
		item{title: "Item 1", description: "Descrpiption item 1"},
		item{title: "Item 2", description: "Description item 2"},
		item{title: "Item 3", description: "Description item 3"},
		item{title: "Item 4", description: "Description item 4"},
		item{title: "About", description: "Contact for Help/Support"},
	}
	// Setup list
	delegate := newItemDelegate(delegateKeys)
	// delegate := list.NewDefaultDelegate() // ← this is the default delegate

	automataMenuItems := list.New(items, delegate, 0, 0)
	automataMenuItems.Title = "Automata"
	automataMenuItems.Styles.Title = titleStyle
	automataMenuItems.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleSpinner,
			listKeys.launchItem,
			listKeys.toggleTitleBar,
			listKeys.toggleStatusBar,
			listKeys.togglePagination,
			listKeys.toggleHelpMenu,
		}
	}

	mainModel := model{
		list:         automataMenuItems,
		keys:         listKeys,
		delegateKeys: delegateKeys,
	}

	aboutModel := about{}
	simpleListModel := listitemmodel{}

	models = []tea.Model{}
	models = append(models, mainModel)
	models = append(models, NewTimer(time.Minute))
	models = append(models, aboutModel)
	models = append(models, simpleListModel)

	return models

}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	return appStyle.Render(m.list.View())

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.toggleSpinner):
			cmd := m.list.ToggleSpinner()
			return m, cmd

		case key.Matches(msg, m.keys.toggleTitleBar):
			v := !m.list.ShowTitle()
			m.list.SetShowTitle(v)
			m.list.SetShowFilter(v)
			m.list.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.keys.toggleStatusBar):
			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.keys.togglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil

		case key.Matches(msg, m.keys.launchItem):
			currentIndex := m.list.Index()
			// This is where we load the components for each page
			if currentIndex == 1 {
				current = timerModel
				return models[current], models[current].Init()

			} else if currentIndex == 4 {
				current = aboutModel
				return models[current], models[current].Init()
			} else if currentIndex == 0 {
				current = simpleListModel
				return models[current], models[current].Init()
			}
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) updateKeybindins() {
	if m.list.SettingFilter() {
		m.keyMap.Enter.SetEnabled(false)
	}
}

func main() {
	models = initMenu()

	p := tea.NewProgram(models[current])
	if err := p.Start(); err != nil {
		fmt.Printf("There was an error: %v", err)
		os.Exit(1)
	}
	//read the config, if the config does not exist, then bail out as we dont know what we need to do
	// config, err := loadConfig()
	// if err != nil {
	// 	fmt.Println("An error was encountered while read config")
	// 	os.Exit(1)
	// }

	// // first download the image from
	// beginDownload(&config)

	// // then begin the process of bringing the machine up
	// bringAMachineUp(&config)

	// // test codes are followings
	// isVangrantInstalled := app.IsVagrantIntsalled()

	// if isVangrantInstalled {
	// 	fmt.Println("Vagrant is installed")
	// 	app.VagrantBoxesList()
	// } else {
	// 	fmt.Println("Vagrant is not installed, please run the setup command again")
	// }
}
