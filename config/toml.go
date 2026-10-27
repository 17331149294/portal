package config

import (
	"strings"
)

// parseTOML is a minimal, self-contained TOML-subset parser.
// It supports the flat structure used by this project's config file:
//   [section]
//   key = "string"
//   key = 123
//   key = true
//   # comment lines
//   key = value  # trailing comment
// It returns a map of "section.key" -> raw value string.
func parseTOML(data []byte) (map[string]string, error) {
	result := make(map[string]string)
	section := ""
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// skip blank lines and full-line comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		// key = value
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		if key == "" {
			continue
		}
		value := stripComment(strings.TrimSpace(line[idx+1:]))
		full := key
		if section != "" {
			full = section + "." + key
		}
		result[full] = value
	}
	return result, nil
}

// stripComment removes a trailing "#" comment, ignoring "#" inside quotes.
func stripComment(s string) string {
	inQuote := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			if !inQuote {
				inQuote = true
			} else {
				inQuote = false
			}
		case '#':
			if !inQuote {
				return strings.TrimSpace(s[:i])
			}
		}
	}
	return s
}

// unquote removes surrounding double quotes if present.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
