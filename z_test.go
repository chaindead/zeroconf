package zerocfg

import (
	"net"
	neturl "net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const name = "sample.name"

type anyFn[T any] func(string, T, string, ...OptNode) *T

func regSource[T any](fn anyFn[T], v T) (reg func() any, value any, source map[string]any) {
	var def T
	reg = func() any {
		return fn(name, def, "desc")
	}

	source = map[string]any{name: v}

	return reg, v, source
}

func Test_ValueOk(t *testing.T) {
	tests := []struct {
		varType string
		init    func() (ptr func() any, val any, src map[string]any)
	}{
		{
			varType: "int",
			init: func() (func() any, any, map[string]any) {
				return regSource(Int, 5)
			},
		},
		{
			varType: "uint",
			init: func() (func() any, any, map[string]any) {
				return regSource(Uint, uint(42))
			},
		},
		{
			varType: "int32",
			init: func() (func() any, any, map[string]any) {
				return regSource(Int32, int32(123))
			},
		},
		{
			varType: "uint32",
			init: func() (func() any, any, map[string]any) {
				return regSource(Uint32, uint32(456))
			},
		},
		{
			varType: "int64",
			init: func() (func() any, any, map[string]any) {
				return regSource(Int64, int64(789))
			},
		},
		{
			varType: "uint64",
			init: func() (func() any, any, map[string]any) {
				return regSource(Uint64, uint64(1011))
			},
		},
		{
			varType: "bool true",
			init: func() (func() any, any, map[string]any) {
				return regSource(Bool, true)
			},
		},
		{
			varType: "bool false",
			init: func() (func() any, any, map[string]any) {
				return regSource(Bool, false)
			},
		},
		{
			varType: "bools",
			init: func() (func() any, any, map[string]any) {
				return regSource(Bools, []bool{true, false, true})
			},
		},
		{
			varType: "ints",
			init: func() (func() any, any, map[string]any) {
				return regSource(Ints, []int{1, 2, 3})
			},
		},
		{
			varType: "string",
			init: func() (func() any, any, map[string]any) {
				return regSource(Str, "value")
			},
		},
		{
			varType: "strings",
			init: func() (func() any, any, map[string]any) {
				return regSource(Strs, []string{"a", "b", "c"})
			},
		},
		{
			varType: "float32",
			init: func() (func() any, any, map[string]any) {
				return regSource(Float32, float32(3.14))
			},
		},
		{
			varType: "floats32",
			init: func() (func() any, any, map[string]any) {
				return regSource(Floats32, []float32{1.1, 2.2, 3.3})
			},
		},
		{
			varType: "float64",
			init: func() (func() any, any, map[string]any) {
				return regSource(Float64, 3.14159265359)
			},
		},
		{
			varType: "floats64",
			init: func() (func() any, any, map[string]any) {
				return regSource(Floats64, []float64{1.1, 2.2, 3.3})
			},
		},
		{
			varType: "duration",
			init: func() (func() any, any, map[string]any) {
				return regSource(Dur, 5*time.Second)
			},
		},
		{
			varType: "durations",
			init: func() (func() any, any, map[string]any) {
				return regSource(Durs, []time.Duration{time.Second, 2 * time.Minute, 3 * time.Hour})
			},
		},
		{
			varType: "ip",
			init: func() (func() any, any, map[string]any) {
				return regSource(ipInternal, net.ParseIP("192.168.1.1"))
			},
		},
		{
			varType: "map",
			init: func() (func() any, any, map[string]any) {
				return regSource(func(name string, value map[string]any, usage string, opts ...OptNode) *map[string]any {
					return Any(name, value, usage, newMapValue, opts...)
				}, map[string]any{"float": 1., "str": "val"})
			},
		},
		{
			varType: "url",
			init: func() (reg func() any, val any, src map[string]any) {
				defaultURLStr := "http://default.com/path"
				valStr := "https://example.com/another?query=1"

				parsedStdURL, err := neturl.Parse(valStr)
				require.NoError(t, err)
				expectedVal := urlValue(*parsedStdURL)

				reg = func() any {
					return URL(name, defaultURLStr, "url description")
				}
				return reg, expectedVal, map[string]any{name: valStr}
			},
		},
		{
			varType: "url empty",
			init: func() (reg func() any, val any, src map[string]any) {
				defaultURLStr := "http://default.com/path"
				valStr := ""

				expectedVal := urlValue{}

				reg = func() any {
					return URL(name, defaultURLStr, "url description")
				}
				return reg, expectedVal, map[string]any{name: valStr}
			},
		},
	}

	dereference := func(t *testing.T, v any) any {
		val := reflect.ValueOf(v)
		require.True(t, val.Kind() == reflect.Ptr, "val must be a pointer")
		elem := val.Elem()
		return elem.Interface()
	}

	for _, tt := range tests {
		t.Run(tt.varType, func(t *testing.T) {
			c = testConfig()

			reg, expected, source := tt.init()
			ptr := reg()

			err := Parse(newMock(source))
			require.NoError(t, err)

			actual := dereference(t, ptr)

			if tt.varType == "url" || tt.varType == "url empty" {
				actualURLValue, okActual := actual.(urlValue)
				require.True(t, okActual, "actual should be urlValue")
				expectedURLValue, okExpected := expected.(urlValue)
				require.True(t, okExpected, "expected should be urlValue")

				actualStdURL := neturl.URL(actualURLValue)
				expectedStdURL := neturl.URL(expectedURLValue)
				require.Equal(t, expectedStdURL.String(), actualStdURL.String())
			} else {
				require.EqualValues(t, expected, actual)
			}

			// check Set and ToString is compatible
			node, ok := c.vs[name]
			require.True(t, ok)

			stringRepresentation := ToString(actual)

			err = node.Value.Set(stringRepresentation)
			require.NoError(t, err)

			updatedActual := dereference(t, ptr)

			if tt.varType == "url" || tt.varType == "url empty" {
				updatedActualURLValue, okUpdated := updatedActual.(urlValue)
				require.True(t, okUpdated)
				actualURLValue, okActual := actual.(urlValue)
				require.True(t, okActual)

				updatedStdURL := neturl.URL(updatedActualURLValue)
				actualStdURL := neturl.URL(actualURLValue)
				require.Equal(t, actualStdURL.String(), updatedStdURL.String())

			} else {
				require.Equal(t, actual, updatedActual)
			}

			// check type name
			cleanVarType := strings.Split(tt.varType, " ")[0]
			require.Equal(t, cleanVarType, node.Value.Type())
		})
	}
}

func Test_ShowIP(t *testing.T) {
	// Save current global config
	savedC := c
	defer func() { c = savedC }()

	// Create fresh config
	c = testConfig()

	// Test IP display in Show()
	_ = IP("test.ip", "192.168.1.1", "test IP address")

	err := Parse()
	require.NoError(t, err)

	output := Show()
	require.Contains(t, output, "192.168.1.1")
	require.NotContains(t, output, "[") // Should not contain byte array representation
}

func Test_IPSlice(t *testing.T) {
	// Save current global config
	savedC := c
	defer func() { c = savedC }()

	// Create fresh config
	c = testConfig()

	// Test IP slice
	ips := IPs("servers", []string{"127.0.0.1", "192.168.1.1"}, "server IPs")

	err := Parse(newMock(map[string]any{
		"servers": `["10.0.0.1", "10.0.0.2"]`,
	}))
	require.NoError(t, err)

	require.Len(t, *ips, 2)
	require.Equal(t, "10.0.0.1", (*ips)[0].String())
	require.Equal(t, "10.0.0.2", (*ips)[1].String())
}
