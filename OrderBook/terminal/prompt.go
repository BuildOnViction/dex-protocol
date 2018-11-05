package terminal

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

type ArgumentHide func(results map[string]string, argument *Argument) bool

type Argument struct {
	Name      string
	Value     string
	AllowEdit bool
	Hide      ArgumentHide
	Validate  promptui.ValidateFunc
	Remember  bool
}

type Command struct {
	Name        string
	Arguments   []Argument
	Description string
}

type CommandsByName []Command

func (c CommandsByName) Len() int {
	return len(c)
}

func (c CommandsByName) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

func (c CommandsByName) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

var (
	promptTpl = &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F449 {{ .Name | cyan }} ({{ .FormatArguments | yellow }})",
		Inactive: "  {{ .Name | cyan }} ({{ .FormatArguments | yellow }})",
		Selected: "\U0001F447 {{ .Name | red | cyan }}",
		Details: `
--------- Command ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}
{{ if .Arguments }}
{{ "Arguments\tDefault Value" | faint }}
{{ .FormatArgumentsWithValue | yellow }}
{{ end }}`,
	}
)

func NewPrompt(label string, size int, commands []Command) *promptui.Select {
	return &promptui.Select{
		Label:     label,
		Items:     commands,
		Templates: promptTpl,
		Size:      size,
		Searcher: func(input string, index int) bool {
			command := commands[index]
			name := strings.Replace(strings.ToLower(command.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		},
		// StartInSearchMode: true,
	}
}

// small struct for config does not need pointer
func (command Command) FormatArguments() string {
	var buffer bytes.Buffer
	for i, argument := range command.Arguments {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(argument.Name)
	}
	return buffer.String()
}

func (command Command) FormatArgumentsWithValue() string {
	var buffer bytes.Buffer

	for _, argument := range command.Arguments {
		buffer.WriteString("  ")
		buffer.WriteString(argument.Name)
		if argument.Value != "" {
			buffer.WriteString("\t  ")
			buffer.WriteString(argument.Value)
		}
		buffer.WriteString("\n")
	}

	return buffer.String()
}

func (command Command) Run() map[string]string {

	validate := func(input string) error {
		if len(input) == 0 {
			return fmt.Errorf("Input is empty")
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	var results = make(map[string]string)
	for _, argument := range command.Arguments {
		// check argument hide function
		if argument.Hide != nil {
			hide := argument.Hide(results, &argument)
			if hide {
				continue
			}
		}

		prompt := promptui.Prompt{
			Label:     argument.Name,
			Default:   argument.Value,
			AllowEdit: argument.AllowEdit,
			Templates: templates,
			Validate: func(input string) error {
				if argument.Validate != nil {
					return argument.Validate(input)
				}
				// default validate
				return validate(input)
			},
		}

		result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
		}

		results[argument.Name] = result

		if argument.Remember {
			argument.Value = result
		}

	}
	return results
}
