package templatecmd

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/sergeyb/aoo/internal/notes"
)

var placeholderPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_-]+)\s*\}\}`)

type Prepared struct {
	Command string
	Values  map[string]string
}

func Prompt(entry notes.Entry, stdin io.Reader, stdout io.Writer, baseValues map[string]string) (Prepared, bool, error) {
	reader := bufio.NewReader(stdin)
	values := make(map[string]string, len(entry.Args)+len(baseValues))
	for key, value := range baseValues {
		values[key] = value
	}

	fmt.Fprintf(stdout, "[template] %s\n", entry.Desc)
	for _, arg := range entry.Args {
		label := arg.Prompt
		if strings.TrimSpace(label) == "" {
			label = arg.Name
		}

		var meta []string
		if strings.TrimSpace(arg.Example) != "" {
			meta = append(meta, "example: "+arg.Example)
		}
		if strings.TrimSpace(arg.Default) != "" {
			meta = append(meta, "default: "+arg.Default)
		}
		if strings.TrimSpace(arg.Description) != "" {
			meta = append(meta, arg.Description)
		}
		if len(meta) > 0 {
			fmt.Fprintf(stdout, "  %s\n", strings.Join(meta, " | "))
		}

		fmt.Fprintf(stdout, "%s: ", label)
		raw, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return Prepared{}, false, err
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			raw = arg.Default
		}
		values[arg.Name] = raw
		if err == io.EOF {
			break
		}
	}

	command, err := Render(entry.Template, values)
	if err != nil {
		return Prepared{}, false, err
	}

	fmt.Fprintf(stdout, "\n[command]\n%s\n", command)
	fmt.Fprint(stdout, "run? [y/N]: ")
	answer, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return Prepared{}, false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		return Prepared{Command: command, Values: values}, false, nil
	}

	return Prepared{Command: command, Values: values}, true, nil
}

func Render(template string, values map[string]string) (string, error) {
	var missing []string
	rendered := placeholderPattern.ReplaceAllStringFunc(template, func(match string) string {
		submatches := placeholderPattern.FindStringSubmatch(match)
		if len(submatches) != 2 {
			return match
		}
		key := submatches[1]
		value, ok := values[key]
		if !ok {
			missing = append(missing, key)
			return match
		}
		return shellQuote(value)
	})

	if len(missing) > 0 {
		return "", fmt.Errorf("missing template args: %s", strings.Join(missing, ", "))
	}

	return rendered, nil
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
