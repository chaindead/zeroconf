package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Parser struct {
	sourceURL **url.URL

	conv    func(any) string
	awaited map[string]bool
}

func New(sourceURL **url.URL) *Parser {
	return &Parser{sourceURL: sourceURL}
}

func (p *Parser) Type() string {
	if p.sourceURL == nil || *(p.sourceURL) == nil {
		return "json"
	}
	return fmt.Sprintf("json[%s]", (*(p.sourceURL)).String())
}

func fetchJSONData(u *url.URL) ([]byte, error) {
	if u == nil {
		return nil, fmt.Errorf("source URL is nil")
	}

	var reader io.ReadCloser
	var err error

	switch u.Scheme {
	case "http", "https":
		resp, httpErr := http.Get(u.String())
		if httpErr != nil {
			return nil, fmt.Errorf("http get failed for %s: %w", u.String(), httpErr)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("http status %s for %s", resp.Status, u.String())
		}
		reader = resp.Body
	case "file":
		path := u.Path
		if u.Host != "" && (os.PathSeparator == '\\' && strings.HasPrefix(path, "/")) {
			path = "//" + u.Host + path
		} else if u.Host != "" {
			path = u.Host + path
		}
		reader, err = os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("file open failed for %s: %w", path, err)
		}
	default:
		if u.Scheme == "" && u.Path != "" {
			reader, err = os.Open(u.Path)
			if err != nil {
				return nil, fmt.Errorf("local file open failed for %s: %w", u.Path, err)
			}
		} else {
			return nil, fmt.Errorf("unsupported URL scheme: %q for %s", u.Scheme, u.String())
		}
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read failed for %s: %w", u.String(), err)
	}
	return data, nil
}

func (p *Parser) Parse(keys map[string]bool, conv func(any) string) (found, unknown map[string]string, err error) {
	if p.sourceURL == nil || *(p.sourceURL) == nil {
		return make(map[string]string), make(map[string]string), nil
	}

	currentURL := *(p.sourceURL)

	data, err := fetchJSONData(currentURL)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch json data from %s: %w", currentURL.String(), err)
	}

	p.conv = conv
	p.awaited = keys

	return p.parse(data)
}

func (p *Parser) parse(data []byte) (found, unknown map[string]string, err error) {
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("unmarshal json: empty json input")
	}
	var settings any

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err = decoder.Decode(&settings); err != nil {
		return nil, nil, fmt.Errorf("unmarshal json: %w", err)
	}

	settingsMap, ok := settings.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("json root is not an object")
	}

	found, unknown = p.flatten(settingsMap)
	return found, unknown, nil
}

func (p *Parser) flatten(settings map[string]any) (found, unknown map[string]string) {
	found, unknown = make(map[string]string), make(map[string]string)
	p.flattenDFS(settings, "", found, unknown)
	return found, unknown
}

func (p *Parser) flattenDFS(m map[string]any, prefix string, found, unknown map[string]string) {
	for k, v := range m {
		if v == nil {
			continue
		}

		newKey := k
		if prefix != "" {
			newKey = prefix + "." + k
		}

		if p.awaited[newKey] {
			found[newKey] = p.conv(v)
			continue
		}

		if subMap, ok := v.(map[string]any); ok {
			p.flattenDFS(subMap, newKey, found, unknown)
			continue
		}

		unknown[newKey] = p.conv(v)
	}
}
