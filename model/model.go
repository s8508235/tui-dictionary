package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaaton/golem/v4"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/s8508235/tui-dictionary/pkg/dictionary"
	"github.com/sirupsen/logrus"
)

type dictionaryState int

const (
	dictionaryInit dictionaryState = iota
	dictionarySearch
	dictionarySelectDef
)

type Dictionary struct {
	Target     textinput.Model
	SearchWord textinput.Model
	// selection
	Choices    []string         // items on the to-do list
	Cursor     int              // which to-do list item our cursor is pointing at
	Selected   map[int]struct{} // which to-do items are selected
	target     string
	searchWord string
	emptyWarn  string
	state      dictionaryState
	// dependencies
	Logger     *logrus.Logger
	OutFile    *os.File
	Lemmatizer *golem.Lemmatizer
	Dictionary dictionary.Interface
	Spinner    *spinner.Spinner
}

func (m Dictionary) Init() tea.Cmd {
	return textinput.Blink
}

func (m Dictionary) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// enter target -> loop (enter word, select definition)
	var cmd tea.Cmd
	switch m.state {
	case dictionaryInit:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				m.state = dictionarySearch
				m.target = m.Target.Value()
				m.Logger.Debugln("select target", m.target)
				file, err := os.OpenFile(filepath.Clean(fmt.Sprintf("%s.txt", m.target)), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
				if err != nil {
					m.Logger.Errorln("Fail to open output file", err)
					return m, tea.Quit
				}
				m.OutFile = file
				m.Target.Blur()
				m.SearchWord.Focus()
				return m, textinput.Blink
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}
		// We handle errors just like any other message
		case error:
			m.Logger.Error(msg)
			return m, tea.Quit
		}
		m.Target, cmd = m.Target.Update(msg)
	case dictionarySearch:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				inputWord := strings.TrimSpace(m.SearchWord.Value())
				m.searchWord = m.Lemmatizer.Lemma(inputWord)
				m.Logger.Debugln("going to search", m.searchWord)
				m.Spinner.Start()
				results, err := m.Dictionary.Search(m.searchWord)
				if err == dictionary.ErrorNoDef {
					fmt.Printf("no definition for: %s\n", m.searchWord)
					shouldCursorReset := m.SearchWord.Reset()
					if shouldCursorReset {
						m.SearchWord.Focus()
					}
					return m, nil
				} else if err != nil {
					m.Logger.Errorln("Search error:", err)
					return m, nil
				}
				m.Spinner.Stop()
				m.state = dictionarySelectDef
				m.SearchWord.Blur()
				m.Choices = results
				return m, nil
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}
		// We handle errors just like any other message
		case error:
			m.Logger.Error(msg)
			return m, tea.Quit
		}
		m.SearchWord, cmd = m.SearchWord.Update(msg)
	case dictionarySelectDef:

		switch msg := msg.(type) {

		// Is it a key press?
		case tea.KeyMsg:

			// Cool, what was the actual key pressed?
			switch msg.String() {
			case "z":
				if len(m.Selected) == 0 {
					m.emptyWarn = "Please at least select one definition"
					return m, nil
				}
				m.emptyWarn = ""
				flushed := make([]string, 0, len(m.Selected))
				for key := range m.Selected {
					flushed = append(flushed, m.Choices[key])
				}
				if err := writeOutput(m.Logger, m.OutFile, m.target, m.searchWord, flushed); err != nil {
					m.Logger.Error(msg)
					return m, tea.Quit
				}
				m.state = dictionarySearch
				shouldCursorReset := m.SearchWord.Reset()
				if shouldCursorReset {
					m.SearchWord.Focus()
				}
				return m, textinput.Blink
			// These keys should exit the program.
			case "ctrl+c", "q":
				return m, tea.Quit

			// The "up" and "k" keys move the cursor up
			case "up", "k":
				if m.Cursor > 0 {
					m.Cursor--
				}

			// The "down" and "j" keys move the cursor down
			case "down", "j":
				if m.Cursor < len(m.Choices)-1 {
					m.Cursor++
				}

			// The "enter" key and the spacebar (a literal space) toggle
			// the selected state for the item that the cursor is pointing at.
			case "enter", " ", "x":
				_, ok := m.Selected[m.Cursor]
				if ok {
					delete(m.Selected, m.Cursor)
				} else {
					m.Selected[m.Cursor] = struct{}{}
				}
			}
		}
	default:
		m.Logger.Warn("????")
	}
	return m, cmd
}

func (m Dictionary) View() string {
	var s string
	switch m.state {
	case dictionaryInit:
		return fmt.Sprintf("Target: %s [generate .txt file to store output]", m.Target.View())
	case dictionarySearch:
		return fmt.Sprintf("Word: %s [Press enter to search, Ctrl+C or Esc to exit]", m.SearchWord.View())
	case dictionarySelectDef:
		s = fmt.Sprintf("Choose one or more definitions for %s:\nPress space, enter or x to select\n", m.searchWord)
		if len(m.emptyWarn) != 0 {
			s += fmt.Sprintf("\n%s\n\n", m.emptyWarn)
		}
		for i, choice := range m.Choices {
			// Is the cursor pointing at this choice?
			cursor := " " // no cursor
			if m.Cursor == i {
				cursor = ">" // cursor!
			}
			// Is this choice selected?
			checked := " " // not selected
			if _, ok := m.Selected[i]; ok {
				checked = "x" // selected!
			}
			// Render the row
			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
		}

		// The footer
		s += "\nPress z to flush to file\nPress q to quit.\n"
	default:
		return "some went wrong"
	}
	return s
}

func writeOutput(logger *logrus.Logger, out *os.File, target, searchWord string, definition []string) error {
	output := fmt.Sprintf("%s\t%s\n", searchWord, strings.Join(definition, ";"))
	if _, err := out.WriteString(output); err != nil {
		logger.Errorln("Fail to write output file though:", err)
		return err
	}
	logger.Debugln("wrote to", out.Name(), "with word:", searchWord, "definition:", strings.Join(definition, ";"))
	return nil
}
