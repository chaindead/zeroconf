package flag

import (
	"os"
	"strings"
)

type Provider struct{}

func New() Provider {
	return Provider{}
}

func (Provider) Type() string {
	return "flag"
}

func (Provider) Provide(awaited map[string]bool, _ func(any) string) (found, unknown map[string]string, err error) {
	args := os.Args[1:]

	found, unknown = parse(awaited, args)
	return
}

func parse(awaited map[string]bool, args []string) (found, unknown map[string]string) {
	found, unknown = make(map[string]string), make(map[string]string)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		var name string
		if strings.HasPrefix(arg, "--") {
			name = arg[2:]
		} else if strings.HasPrefix(arg, "-") {
			name = arg[1:]
		}

		if name == "" {
			continue
		}

		// Support the --name=value GNU form: split the value off the flag
		// name so it is matched against the awaited key and not dropped.
		var inlineValue string
		if idx := strings.IndexByte(name, '='); idx >= 0 {
			inlineValue = name[idx+1:]
			name = name[:idx]
		}

		var value string
		if inlineValue != "" {
			value = inlineValue
		} else if i+1 < len(args) && len(args[i+1]) > 0 && args[i+1][0] != '-' {
			value = args[i+1]
			i++
		}

		if _, ok := awaited[name]; ok {
			found[name] = value
		} else {
			unknown[name] = value
		}
	}

	return
}
