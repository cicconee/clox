package env

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Loader struct {
	ReadCloser io.ReadCloser
}

func NewFileLoader(filepath string) (*Loader, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %v", filepath, err)
	}

	return &Loader{ReadCloser: file}, nil
}

// Load will read the content in ReadCloser line by line and set the environment
// variables. If a syntax error is found an error will be returned including the
// line number in filepath.
func (l *Loader) Load() error {
	scanner := bufio.NewScanner(l.ReadCloser)
	lineNumber := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		kv, ok := parseLine(line)
		if !ok {
			return fmt.Errorf("syntax error at line %d: %s", lineNumber, line)
		}

		err := os.Setenv(kv.key, kv.value)
		if err != nil {
			return fmt.Errorf("setting key %q value %q: %v", kv.key, kv.value, err)
		}
	}

	return l.ReadCloser.Close()
}

// keyValue represents a environment variable key-value pair.
type keyValue struct {
	key   string
	value string
}

// parseLine will parse a string to extract the environment variable. It expects
// the line to be formatted as KEY=VALUE. White space will be trimmed on both
// the key and value.
//
// If the line is parsed successfully a nonempty keyvalue and
// true will be returned, otherwise an empty keyvalue and false is returned.
func parseLine(line string) (keyValue, bool) {
	parts := strings.Split(line, "=")
	if len(parts) != 2 {
		return keyValue{}, false
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == "" || value == "" {
		return keyValue{}, false
	}

	return keyValue{key, value}, true
}
