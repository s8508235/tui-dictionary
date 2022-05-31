package model

import (
	"fmt"
	"os"
	"strings"

	"github.com/aaaton/golem/v4"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	cursor   int              // which to-do list item our cursor is pointing at
	Selected map[int]struct{} // which to-do items are selected
	// internal
	searchWord string
	warnMsg    string
	state      dictionaryState
	err        error
	height     int
	width      int
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
	switch msg := msg.(type) {
	// init/update for width/height
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		return m, nil
	}
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
					if len(inputWord) != 0 {
						m.warnMsg = fmt.Sprintf("wrong format of input: %s", inputWord)
					} else {
						m.warnMsg = "empty input"
					}
					shouldCursorReset := m.SearchWord.Reset()
					if shouldCursorReset {
						m.SearchWord.Focus()
					}
					return m, textinput.Blink
				default:
					m.err = fmt.Errorf("fail to ask: %w", err)
					return m, tea.Quit
				}
				m.searchWord = m.Lemmatizer.Lemma(inputWord)
				m.warnMsg = ""
				m.Logger.Debugln("going to search", m.searchWord)
				// if stuck, no way to return. have to be more responsive
				/* test result
				results := []string{
					"a",
					"b",
					"c",
					"d",
					"e",
					"f",
					"g",
					"h",
					"i",
					"j",
					"k",
					"l",
					"m",
					"n",
					"o",
					"p",
					"q",
					"r",
					"s",
					"t",
					"u",
					"",
				}
				*/
				m.Spinner.Start()
				results, err := m.Dictionary.Search(m.searchWord)
				if err == dictionary.ErrorNoDef {
					m.warnMsg = fmt.Sprintf("no definition for: %s\n", m.searchWord)
					// reset search state
					shouldCursorReset := m.SearchWord.Reset()
					if shouldCursorReset {
						m.SearchWord.Focus()
					}
					m.Spinner.Stop()
					return m, textinput.Blink
				} else if err != nil {
					m.err = fmt.Errorf("fail to search: %w", err)
					m.Spinner.Stop()
					return m, tea.Quit
				}
				m.Spinner.Stop()
				// go to selectDef state
				m.state = dictionarySelectDef
				m.SearchWord.Blur()
				m.Choices = results
				m.cursor = 0
				return m, nil
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}
		// We handle errors just like any other message
		case error:
			m.err = msg
			return m, tea.Quit
		}
		m.SearchWord, cmd = m.SearchWord.Update(msg)
	case dictionarySelectDef:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "f", "F", "ctrl+s", "ctrl+S":
				if len(m.Selected) == 0 {
					m.warnMsg = "Please at least select one definition"
					return m, nil
				}
				flushed := make([]string, 0, len(m.Selected))
				for key := range m.Selected {
					flushed = append(flushed, m.Choices[key])
				}
				if err := writeOutput(m.Logger, m.OutFile, m.searchWord, flushed); err != nil {
					m.err = fmt.Errorf("fail to write output file: %w", err)
					return m, tea.Quit
				}
				// back to search state
				return m.backToSearch(), textinput.Blink
			// These keys should exit the program.
			case "ctrl+c", "ctrl+C":
				return m, tea.Quit
			case "up", "w", "W":
				m.cursor = (m.cursor - 1 + len(m.Choices)) % len(m.Choices)
			case "down", "s", "S":
				m.cursor = (m.cursor + 1 + len(m.Choices)) % len(m.Choices)
			case "enter", " ", "x", "X":
				_, ok := m.Selected[m.cursor]
				if ok {
					delete(m.Selected, m.cursor)
				} else {
					m.Selected[m.cursor] = struct{}{}
				}
			case "q", "Q":
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
	// TODO: too long for width and height
	var s string
	switch m.state {
	case dictionarySearch:
		s += fmt.Sprintf("Target: %s\nWord: %s [Press enter to search, Ctrl+C or Esc to exit]", m.Target, m.SearchWord.View())
		if len(m.warnMsg) != 0 {
			s += fmt.Sprintf("\n\033[31m%s\033[0m\n", m.warnMsg)
		}
	case dictionarySelectDef:
		header := fmt.Sprintf("Target: %s\n", m.Target)
		header += fmt.Sprintf("Choose one or more definitions for \033[32m%s\033[0m:\n\n", m.searchWord)
		if len(m.warnMsg) != 0 {
			header += fmt.Sprintf("\033[31m%s\033[0m\n\n", m.warnMsg)
		}
		footer := "\nPress space, enter or x to select\nPress q to skip\n"
		footer += "Press f or Ctrl + s to flush\nPress Ctrl + c to quit.\n"
		remainHeight := lipgloss.Height(header) + lipgloss.Height(footer) + 1
		currentHeight := remainHeight
		var content string
		for i, choice := range m.Choices {
			if m.cursor-i > m.height-remainHeight {
				continue
			}
			// Is the cursor pointing at this choice?
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor!
			}
			// Is this choice selected?
			checked := " " // not selected
			if _, ok := m.Selected[i]; ok {
				checked = "x" // selected!
			}
			// Render the row
			line := fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
			// longer than terminal
			if width := lipgloss.Width(line); width > m.width+1 {
				// since we replace all \s with space when search
				// for curr, end := 0, m.width; curr < width; {
				// 	if end < width {
				// 		content += fmt.Sprintf("%s\n", line[curr:end])
				// 		currentHeight++
				// 		if currentHeight > m.height {
				// 			break
				// 		}
				// 		end += m.width
				// 	} else {
				// 		content += fmt.Sprintf("%s\n", line[curr:])
				// 		currentHeight++
				// 		if currentHeight > m.height {
				// 			break
				// 		}
				// 	}
				// 	curr += m.width
				// }
				content += fmt.Sprintf("%s...\n", line[:m.width-3])
				currentHeight++
				if currentHeight > m.height {
					break
				}
			} else {
				content += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
				currentHeight++
				if currentHeight > m.height {
					break
				}
			}
		}
		s = fmt.Sprintf("%s%s%s", header, content, footer)
	default:
		return "some went wrong"
	}
	return s
}

func (m Dictionary) backToSearch() Dictionary {
	m.warnMsg = ""
	m.Selected = make(map[int]struct{})
	m.Choices = make([]string, 0)
	m.cursor = 0
	m.state = dictionarySearch
	shouldCursorReset := m.SearchWord.Reset()
	if shouldCursorReset {
		m.SearchWord.Focus()
	}
	return m
}

func (m Dictionary) GetError() error {
	return m.err
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
