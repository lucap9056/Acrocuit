package models

import (
	"strings"
)

type columnScanner struct {
	str     string
	current string
}

func newColumnScanner(s string) columnScanner {
	return columnScanner{str: s}
}

func (s *columnScanner) Next() bool {
	for len(s.str) > 0 {
		var col string
		if idx := strings.IndexByte(s.str, ','); idx >= 0 {
			col = strings.TrimSpace(s.str[:idx])
			s.str = s.str[idx+1:]
		} else {
			col = strings.TrimSpace(s.str)
			s.str = ""
		}

		if col != "" {
			s.current = col
			return true
		}
	}
	s.current = ""
	return false
}

func (s *columnScanner) Text() string {
	return s.current
}
