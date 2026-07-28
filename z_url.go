package zerocfg

import (
	"fmt"
	"net/url"
)

type urlValue url.URL

func newURLValue(val urlValue, p *urlValue) Value {
	return p
}

func (u *urlValue) Set(s string) error {
	if s == "" {
		*u = urlValue{}
		return nil
	}
	parsedURL, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("parsing URL %q: %w", s, err)
	}
	if parsedURL == nil {
		*u = urlValue{}
		return nil
	}
	*u = urlValue(*parsedURL)
	return nil
}

func (u *urlValue) Type() string {
	return "url"
}

func (u *urlValue) String() string {
	if u == nil {
		return ""
	}
	tempOriginalURL := url.URL(*u)
	return tempOriginalURL.String()
}

func URL(name string, defValStr string, desc string, opts ...OptNode) *urlValue {
	var initialStruct urlValue

	if defValStr != "" {
		parsedURL, err := url.Parse(defValStr)
		if err != nil {
			panic(fmt.Sprintf("invalid default URL value %q for %q: %v", defValStr, name, err))
		}
		if parsedURL != nil {
			initialStruct = urlValue(*parsedURL)
		}
	}
	return Any(name, initialStruct, desc, newURLValue, opts...)
}
