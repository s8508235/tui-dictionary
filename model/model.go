package model

import (
	"fmt"
	"os"
	"strings"

	"github.com/aaaton/golem/v4"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/s8508235/tui-dictionary/pkg/dictionary"
	"github.com/s8508235/tui-dictionary/pkg/tools"
	"github.com/sirupsen/logrus"
)

type dictionaryState int

const (
	dictionarySearch dictionaryState = iota
	dictionarySelectDef
)

type Dictionary struct {
	Target     string
	SearchWord textinput.Model
	// selection
	Choices  []string         // items on the to-do list
	Cursor   int              // which to-do list item our cursor is pointing at
	Selected map[int]struct{} // which to-do items are selected
	// internal
	searchWord string
	warnMsg    string
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
	var cmd tea.Cmd
	switch m.state {
	case dictionarySearch:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				inputWord := strings.TrimSpace(m.SearchWord.Value())
				err := tools.WordValidate(inputWord)
				switch err {
				case nil:
				case tools.ErrWord:
					m.warnMsg = fmt.Sprintf("wrong format of input: %s", inputWord)
					shouldCursorReset := m.SearchWord.Reset()
					if shouldCursorReset {
						m.SearchWord.Focus()
					}
					return m, textinput.Blink
				default:
					m.Logger.Errorln("Fail to ask:", err)
					return m, tea.Quit
				}
				m.searchWord = m.Lemmatizer.Lemma(inputWord)
				m.Logger.Debugln("going to search", m.searchWord)
				m.Spinner.Start()
				results, err := m.Dictionary.Search(m.searchWord)
				if err == dictionary.ErrorNoDef {
					m.Logger.Debugf("no definition for: %s\n", m.searchWord)
					// reset search state
					shouldCursorReset := m.SearchWord.Reset()
					if shouldCursorReset {
						m.SearchWord.Focus()
					}
					m.Spinner.Stop()
					return m, textinput.Blink
				} else if err != nil {
					m.Logger.Errorln("Search error:", err)
					m.Spinner.Stop()
					return m, tea.Quit
				}
				m.Spinner.Stop()
				// go to selectDef state
				m.state = dictionarySelectDef
				m.SearchWord.Blur()
				m.Choices = results
				m.Cursor = 0
				m.warnMsg = ""
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
		case tea.KeyMsg:
			switch msg.String() {
			case "z", "Z":
				if len(m.Selected) == 0 {
					m.warnMsg = "Please at least select one definition"
					return m, nil
				}
				flushed := make([]string, 0, len(m.Selected))
				for key := range m.Selected {
					flushed = append(flushed, m.Choices[key])
				}
				if err := writeOutput(m.Logger, m.OutFile, m.searchWord, flushed); err != nil {
					m.Logger.Error(msg)
					return m, tea.Quit
				}
				// back to search state
				return m.backToSearch(), textinput.Blink
			// These keys should exit the program.
			case "ctrl+c", "q", "Q":
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
			case "enter", " ", "x", "X":
				_, ok := m.Selected[m.Cursor]
				if ok {
					delete(m.Selected, m.Cursor)
				} else {
					m.Selected[m.Cursor] = struct{}{}
				}
			case "s", "S":
				// back to search state
				return m.backToSearch(), textinput.Blink
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
	case dictionarySearch:
		if len(m.warnMsg) != 0 {
			s += fmt.Sprintf("\033[31m%s\033[0m\n\n", m.warnMsg)
		}
		s += fmt.Sprintf("Target: %s\nWord: %s [Press enter to search, Ctrl+C or Esc to exit]", m.Target, m.SearchWord.View())
	case dictionarySelectDef:
		s += fmt.Sprintf("Target: %s\n", m.Target)
		s += fmt.Sprintf("Choose one or more definitions for \033[32m%s\033[0m:\n\n", m.searchWord)
		if len(m.warnMsg) != 0 {
			s += fmt.Sprintf("\033[31m%s\033[0m\n\n", m.warnMsg)
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
		s += "\nPress space, enter or x to select\nPress s to skip\n"
		// The footer
		s += "Press z to flush\nPress q to quit.\n"
	default:
		return "some went wrong"
	}
	return s
}

func (m Dictionary) backToSearch() Dictionary {
	m.warnMsg = ""
	m.Selected = make(map[int]struct{})
	m.Choices = make([]string, 0)
	m.Cursor = 0
	m.state = dictionarySearch
	shouldCursorReset := m.SearchWord.Reset()
	if shouldCursorReset {
		m.SearchWord.Focus()
	}
	return m
}

func writeOutput(logger *logrus.Logger, out *os.File, searchWord string, definition []string) error {
	output := fmt.Sprintf("%s\t%s\n", searchWord, strings.Join(definition, ";"))
	if _, err := out.WriteString(output); err != nil {
		logger.Errorln("Fail to write output file though:", err)
		return err
	}
	logger.Debugln("wrote to", out.Name(), "with word:", searchWord, "definition:", strings.Join(definition, ";"))
	return nil
}
