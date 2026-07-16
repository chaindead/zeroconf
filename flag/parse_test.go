package flag_test

import (
	"os"
	"testing"

	zfg "github.com/chaindead/zerocfg"
	"github.com/chaindead/zerocfg/flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		awaited map[string]bool
		found   map[string]string
		unknown map[string]string
	}{
		{
			name:    "single short flag with value",
			args:    []string{"-n1", "v1", "-n2", "v2"},
			awaited: map[string]bool{"n1": true},
			found:   map[string]string{"n1": "v1"},
			unknown: map[string]string{"n2": "v2"},
		},
		{
			name:    "single long flag with value",
			args:    []string{"--name", "value"},
			awaited: map[string]bool{"name": true},
			found:   map[string]string{"name": "value"},
		},
		{
			name: "multiple flags with values",
			args: []string{"-host", "localhost", "--port", "5432", "-user", "admin"},
			awaited: map[string]bool{
				"host": true,
				"port": true,
				"user": true,
			},
			found: map[string]string{
				"host": "localhost",
				"port": "5432",
				"user": "admin",
			},
		},
		{
			name:    "flag without value",
			args:    []string{"-debug"},
			awaited: map[string]bool{"debug": true},
			found:   map[string]string{"debug": ""},
		},
		{
			name: "mixed flags with and without values",
			args: []string{"-verbose", "--config", "config.yaml", "-force"},
			awaited: map[string]bool{
				"verbose": true,
				"config":  true,
			},
			found: map[string]string{
				"verbose": "",
				"config":  "config.yaml",
			},
			unknown: map[string]string{
				"force": "",
			},
		},
		{
			name: "ignore non-flag arguments",
			args: []string{"positional", "-name", "value", "another", "-debug"},
			unknown: map[string]string{
				"name":  "value",
				"debug": "",
			},
		},
		{
			name:    "empty args",
			args:    []string{},
		},
		{
			name:    "long flag with inline equals value",
			args:    []string{"--name=value", "--other=foo"},
			awaited: map[string]bool{"name": true, "other": true},
			found:   map[string]string{"name": "value", "other": "foo"},
			unknown: map[string]string{},
		},
		{
			name:    "mixed space and inline equals values",
			args:    []string{"--config", "config.yaml", "--port=5432"},
			awaited: map[string]bool{"config": true, "port": true},
			found:   map[string]string{"config": "config.yaml", "port": "5432"},
			unknown: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.found == nil {
				tt.found = map[string]string{}
			}
			if tt.unknown == nil {
				tt.unknown = map[string]string{}
			}

			p := flag.New()
			os.Args = append([]string{"program"}, tt.args...)

			found, unknown, err := p.Provide(tt.awaited, zfg.ToString)
			require.NoError(t, err)

			assert.Equal(t, tt.found, found)
			assert.Equal(t, tt.unknown, unknown)
		})
	}
}
