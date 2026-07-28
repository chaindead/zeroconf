# Zero Effort Configuration

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go?tab=readme-ov-file#configuration)
[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/chaindead/zerocfg)
[![Go Report Card](https://goreportcard.com/badge/github.com/chaindead/zerocfg)](https://goreportcard.com/report/github.com/chaindead/zerocfg)
[![Codecov](https://codecov.io/gh/chaindead/zerocfg/branch/main/graph/badge.svg)](https://codecov.io/gh/chaindead/zerocfg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![visitors](https://api.visitorbadge.io/api/visitors?path=github.com%2Fchaindead%2Fzerocfg&label=Visitors&countColor=%23263759&style=plastic&labelStyle=none)](https://visitorbadge.io/status?path=github.com%2Fchaindead%2Fzerocfg)

I've always loved the elegance of Go's flag package - how clean and straightforward it is to define and use configuration options. While working on various Go projects, I found myself wanting that same simplicity but with support for YAML configs. I couldn't find anything that preserved this paradigm, so I built zerocfg.

- 🛠️ Simple and flexible API inspired by `flag` package
- 🍳 Boilerplate usage prohibited by design
- 🚦 Early detection of mistyped config keys
- ✨ Multiple configuration sources with priority-based value resolution
- 🕵️‍♂️ Render running configuration with secret protection
- 🧩 Custom option types and providers are supported

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
  - [Options naming](#options-naming)
  - [Restrictions](#restrictions)
  - [Unknown values](#unknown-values)
  - [Complex Types as string](#complex-types-as-string)
- [Configuration Sources](#configuration-sources)
  - [Command-line Arguments](#command-line-arguments)
  - [Environment Variables](#environment-variables)
  - [YAML Source](#yaml-source)
- [Advanced Usage](#advanced-usage)
  - [Value Representation](#value-representation)
  - [Custom Options](#custom-options)
  - [Custom Providers](#custom-providers)

## Installation

```bash
go get -u github.com/chaindead/zerocfg
```

## Quick Start

```go
package main

import (
    "fmt"

    zfg "github.com/chaindead/zerocfg"
    "github.com/chaindead/zerocfg/env"
    "github.com/chaindead/zerocfg/yaml"
)

var (
    // Configuration variables
    path     = zfg.Str("config.path", "", "path to yaml conf file", zfg.Alias("c"))
    ip       = zfg.IP("db.ip", "127.0.0.1", "database location")
    port     = zfg.Uint("db.port", 5678, "database port")
    username = zfg.Str("db.user", "guest", "user of database")
    password = zfg.Str("db.password", "qwerty", "password for user", zfg.Secret())
)

func main() {
    // Initialize configuration with multiple sources
    err := zfg.Parse(
        env.New(),
        yaml.New(path),
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("Connect to %s:%d creds=%s:%s\n", *ip, *port, *username, *password)
    // OUTPUT: Connect to 127.0.0.1:5678 creds=guest:qwerty

    fmt.Println(zfg.Show())
    // CMD: go run ./... -c test.yaml
    // OUTPUT:
    //  config.path = test.yaml      (path to yaml conf file)
    //  db.ip       = 127.0.0.1      (database location)
    //  db.password = <secret>       (password for user)
    //  db.port     = 5678           (database port)
    //  db.user     = guest          (user of database)
}
```

## Usage

### Options naming

- Dots (`.`) are used as separators for hierarchical options.
- Option subnames preferred separation is camelCase, underscore (`_`), and dash (`-`) styles.

**Example:**

```go
zfg.Str("groupOptions.thisOption", "", "camelCase usage")
zfg.Str("group_options.this_option", "", "underscore usage")
zfg.Str("group-options.this-option", "", "dash usage")
```

### Restrictions

- Options are registered at import time. Dynamic (runtime) option registration is not supported

    ```go
    // internal/db/client.go
    package db

    import zfg "github.com/chaindead/zerocfg"

    // good: options registered at import
    var dbHost = zfg.Str("db.host", "localhost", "called on import")

    // bad: dynamic registration
    func AddOption() {
        zfg.Str("db.dynamic", "", "not supported")
    }
    ```

- No key duplication is allowed. Each option key must be unique to ensure a single source of truth and avoid boilerplate
- Simultaneous use of keys and sub-keys (e.g., `map` and `map.value`) are not allowed

### Unknown values

If `zfg.Parse` encounters an unknown value (e.g. variable not registered as an option), it returns an error. 
This helps avoid boilerplate and ensures only declared options are used. 

But you can ignore unknown values if desired.

```go
err := zfg.Parse(
    env.New(),
    yaml.New(path),
)
if u, ok := zfg.IsUnknown(err); !ok {
    panic(err)
} else {
    // u is map <source_name> to slice of unknown keys
    fmt.Println(u)
}
```

> `env` source does not trigger unknown options to avoid false positives.

### Complex Types as string

- Base values converted via `fmt.Sprint("%v")`
- If a type has a `String()` method, it is used for string conversion (e.g., `time.Duration`).
- Otherwise, JSON representation is used for complex types (e.g., slices, maps).

> For converting any value to string, `zfg.ToString` is used internally.

```go
var (
    _ = zfg.Dur("timeout", 5*time.Second, "duration via fmt.Stringer interface")
    _ = zfg.Floats64("floats", nil, "list via json")
)

func main() {
    _ = zfg.Parse()

    fmt.Printf(zfg.Show())
    // CMD: go run ./... --timeout 10s --floats '[1.1, 2.2, 3.3]'
    // OUTPUT:
    //   floats  = [1.1,2.2,3.3] (list via json)
    //   timeout = 10s           (duration via fmt.Stringer interface)
}

```

## Configuration Sources

The configuration system follows a strict priority hierarchy:

1. Command-line flags (always highest priority, enabled by default)
2. Optional providers in order of addition (first added = higher priority)
3. Default values (lowest priority)

For example, if you initialize configuration like this:
```go
zfg.Parse(
    env.New(),      // Second highest priority (after cli flags)
    yaml.New(path), // Third highest priority
)
```

The final value resolution order will be:
1. Command-line flags (if provided)
2. Providers from arguments of `zfg.Parse` in same order as it is passed.
3. Default values

Important notes:
- Lower priority sources cannot override values from higher priority sources
- All providers except flags are optional
- Provider priority is determined by the order in `Parse()` function
- Values not found in higher priority sources fall back to lower priority sources

### Command-line Arguments

- The flag source is enabled by default and always has the highest priority
- You can define configuration options with aliases for convenient CLI usage
- Values are passed as space-separated arguments (no `=` allowed)
- Both single dash (`-`) and double dash (`--`) prefixes are supported for flags and their aliases

**Example:**

```go
path := zfg.Str("config.path", "", "path to yaml conf file", zfg.Alias("c"))
```

You can run your application with:

```
go run ./... -c test.yaml
# or
go run ./... --config.path test.yaml
```

In both cases, the value `test.yaml` will be assigned to `config.path`.

### Environment Variables

Environment variables are automatically transformed from the configuration key format:

| Config Key | Environment Variable | Note |
|------------|---------------------|------|
| db.user | DB_USER | Basic transformation |
| app.api.key | APP_API_KEY | Multi-level path |
| camelCase.value | CAMELCASE_VALUE | CamelCase handling |
| api-key.secret | APIKEY_SECRET | Dashes removed |
| under_score.value | UNDERSCORE_VALUE | Underscores removed |


The transformation rules:
1. Remove special characters (except letters, digits, and dots)
2. Replace dots with underscores
3. Convert to uppercase

**Example:**

```go
import (
    "fmt"
    zfg "github.com/chaindead/zerocfg"
    "github.com/chaindead/zerocfg/env"
)
var dbUser = zfg.Str("db.user", "", "database's username")

func main() {
    _ = zfg.Parse(
        env.New(),
    )
    fmt.Printf("DB user: %s", *dbUser)
}
```

When you run, `dbUser` will be set to `admin`.

```bash
DB_USER=admin go run main.go
# OUTPUT: DB user: admin
```

### YAML Source

- Options use dotted paths to map to YAML keys, supporting hierarchical configuration.
- All naming styles are supported and mapped to YAML keys as written.

**Example YAML file:**

```yaml
group:
  option: "foo"

numbers:
  - 1
  - 2
  - 3

limits:
  max: 10
  min: 1
```

### JSON Source (via URL)

The JSON parser loads configuration from a URL, supporting `file://`, `http://`, and `https://` schemes. Schemeless paths are treated as local files.

First, define a URL option:

```go
var jsonConfigURL = zfg.URL("json.config", "file:///path/to/your/config.json", "URL for JSON configuration")
// Or for an optional remote config:
// var remoteJsonConfigURL = zfg.URL("json.remote", "", "Optional remote JSON config URL (http/https)")
```

**Example Go config:**

```go
zfg.Str("group.option", "", "hierarchical usage")
zfg.Ints("numbers", nil, "slice of server configs")
zfg.Map("limits", nil, "map of limits")
```
Then, add it to `zerocfg.Parse`:

```go
err := zerocfg.Parse(
    // ... other providers
    json.New(jsonConfigURL),
)
```

If the `json.config` URL option is not provided (or its default is empty), the JSON provider will be skipped without error.

## Advanced Usage

### Value Representation

> [!IMPORTANT]
> Read this section before implementing custom options or providers.
> - All supported option values must have a string representation
> - Conversion to string is performed using `zfg.ToString`
> - Types must implement `Set(string)`; the string passed is produced by `ToString` and parsing must be compatible
> - Providers return `map[string]string` where values are produced by the `conv` function  argument in the provider interface (internally `zfg.ToString` is used)

### Custom Options

You can define your own option types by implementing the `Value` interface and registering them via `Any` function.
Methods `Set` and `String` should be compatible.

```go
// Custom type
type MyType struct{ V string }

func newValue(val MyType, p *MyType) zfg.Value {
    *p = val
    return p
}

func (m *MyType) Set(s string) error { m.V = s; return nil }
func (m *MyType) Type() string       { return "custom" }
func (m *MyType) String() string { return m.V }

func Custom(name string, defVal MyType, desc string, opts ...zfg.OptNode) *MyType {
     return zfg.Any(name, defVal, desc, newValue, opts...)
}

// Register custom option
var myOpt = Custom("custom.opt", MyType{"default"}, "custom option")
```

### Custom Providers

You can add your own configuration sources by implementing the `Provider` interface.

- If `awaited[name] == true`, the name is an option
- If `awaited[name] == false`, the name is an alias

```go
type MyProvider struct{}

func (p *MyProvider) Type() string { return "my" }
func (p *MyProvider) Provide(awaited map[string]bool, conv func(any) string) (map[string]string, map[string]string, error) {
    found := map[string]string{}
    unknown := map[string]string{}
    // ... fill found/unknown based on awaited ...
    return found, unknown, nil
}

// Usage
zfg.Parse(&MyProvider{})
```

## Documentation

For detailed documentation and advanced usage examples, visit our [Godoc page](https://godoc.org/github.com/chaindead/zerocfg).

## Star History

<a href="https://www.star-history.com/#chaindead/zerocfg&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=chaindead/zerocfg&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=chaindead/zerocfg&type=Date" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=chaindead/zerocfg&type=Date" />
 </picture>
</a>

## License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/chaindead/zerocfg/blob/main/LICENCE) file for details.