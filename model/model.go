package model

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aaaton/golem/v4"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/s8508235/tui-dictionary/pkg/dictionary"
	"github.com/s8508235/tui-dictionary/pkg/tools"
	"github.com/sirupsen/logrus"
)

type dictionaryState int
type dictionaryResult []string

const (
	dictionarySearchStart dictionaryState = iota
	dictionarySearching
	dictionarySelectDef
)

type Dictionary struct {
	Target     string
	SearchWord textinput.Model
	Spinner    spinner.Model
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
	Out        io.Writer
	Lemmatizer *golem.Lemmatizer
	Dictionary dictionary.Interface
}

func (m Dictionary) Init() tea.Cmd {
	return textinput.Blink
}
func (m Dictionary) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	// init/update for width/height
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		return m, nil
	}
	switch m.state {
	case dictionarySearchStart:
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

				// go to selectDef state
				m.state = dictionarySearching
				m.SearchWord.Blur()
				return m, tea.Batch(m.Spinner.Tick, m.wordSearch())
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}
		// We handle errors just like any other message
		case error:
			m.err = msg
			return m, tea.Quit
		}
		m.SearchWord, cmd = m.SearchWord.Update(msg)
		return m, cmd
	case dictionarySearching:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "Q":
				// back to search state
				return m.backToSearch(), textinput.Blink
			case "ctrl+c", "ctrl+C":
				return m, tea.Quit
			}
		case dictionaryResult:
			m.Choices = []string(msg)
			m.cursor = 0
			m.state = dictionarySelectDef
			return m, nil
		case error:
			if !errors.Is(msg, dictionary.ErrorNoDef) {
				m.err = msg
				return m, tea.Quit
			}
			m.warnMsg = dictionary.ErrorNoDef.Error()
			shouldCursorReset := m.SearchWord.Reset()
			if shouldCursorReset {
				m.SearchWord.Focus()
			}
			m.state = dictionarySearchStart
			return m, textinput.Blink
		default:
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}
	case dictionarySelectDef:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "c", "C":
				m.warnMsg = ""
				return m, nil
			case "f", "F", "ctrl+s", "ctrl+S":
				if len(m.Selected) == 0 {
					m.warnMsg = "Please at least select one definition"
					return m, nil
				}
				flushed := make([]string, 0, len(m.Selected))
				for key := range m.Selected {
					flushed = append(flushed, m.Choices[key])
				}
				if err := writeOutput(m.Logger, m.Out, m.searchWord, flushed); err != nil {
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
		m.err = errors.New("unreachable")
		return m, tea.Quit
	}
	return m, nil
}

func (m Dictionary) View() string {
	// TODO: too long for width and height
	var s string
	switch m.state {
	case dictionarySearchStart:
		s += fmt.Sprintf("Target: %s\nWord: %s [Press enter to search, Ctrl+C or Esc to exit]", m.Target, m.SearchWord.View())
		if len(m.warnMsg) != 0 {
			s += fmt.Sprintf("\n\033[31m%s\033[0m\n", m.warnMsg)
		}
	case dictionarySearching:
		s = fmt.Sprintf("Target: %s\n", m.Target)
		if len(m.warnMsg) != 0 {
			s += fmt.Sprintf("\n\033[31m%s\033[0m\n", m.warnMsg)
		}
		s += fmt.Sprintf("%s\nEnter q to cancel or Ctrl+C to exit", m.Spinner.View())
	case dictionarySelectDef:
		header := fmt.Sprintf("Target: %s\n", m.Target)
		header += fmt.Sprintf("There are \033[92m%d\033[0m definitions, please choose one or more definitions for \033[92m%s\033[0m:\n\n", len(m.Choices), m.searchWord)
		if len(m.warnMsg) != 0 {
			header += fmt.Sprintf("\033[31m%s\033[0m\n\n", m.warnMsg)
		}
		footer := "\nPress space, enter or x to select\nPress q to skip\n"
		footer += "Press f or Ctrl + s to flush\nPress Ctrl + c to quit."
		remainHeight := lipgloss.Height(header) + lipgloss.Height(footer)
		pageLineCount := m.height - remainHeight + 1
		if pageLineCount < 1 {
			return "too small to show content"
		}
		currentPage := m.cursor / pageLineCount
		currentHeight := remainHeight
		var remainder int
		if len(m.Choices)%pageLineCount != 0 {
			remainder = 1
		}
		footer = fmt.Sprintf("\033[38:2:255:165:0mpage: %2d / %2d\033[0m", currentPage+1, len(m.Choices)/pageLineCount+(remainder)) + footer
		var content string
		for i, choice := range m.Choices {
			if currentPage*pageLineCount > i || (currentPage+1)*pageLineCount <= i {
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
			line := fmt.Sprintf("%s %2d [%s] %s\n", cursor, i+1, checked, choice)
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
				content += fmt.Sprintf("%s...\n", line[:m.width-4])
				currentHeight++
				if currentHeight > m.height {
					break
				}
			} else {
				// content += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
				content += line
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
	m.state = dictionarySearchStart
	shouldCursorReset := m.SearchWord.Reset()
	if shouldCursorReset {
		m.SearchWord.Focus()
	}
	return m
}

func (m Dictionary) GetError() error {
	return m.err
}

func (m Dictionary) wordSearch() tea.Cmd {
	return func() tea.Msg {
		results, err := m.Dictionary.Search(m.searchWord)
		if err != nil {
			return err
		}
		return dictionaryResult(results)
	}
}

func writeOutput(logger *logrus.Logger, out io.Writer, searchWord string, definition []string) error {
	var buf bytes.Buffer
	if _, err := buf.WriteString(searchWord); err != nil {
		logger.Errorln("Fail to write:", err)
		return err
	}
	if _, err := buf.WriteRune('\t'); err != nil {
		logger.Errorln("Fail to write:", err)
		return err
	}
	if _, err := buf.WriteString(strings.Join(definition, ";")); err != nil {
		logger.Errorln("Fail to write:", err)
		return err
	}
	if _, err := buf.WriteRune('\n'); err != nil {
		logger.Errorln("Fail to write:", err)
		return err
	}
	if _, err := out.Write(buf.Bytes()); err != nil {
		logger.Errorln("Fail to write output file though:", err)
		return err
	}
	logger.Debugln("word:", searchWord, "definition:", strings.Join(definition, ";"))
	return nil
}
