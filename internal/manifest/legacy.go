package manifest

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

func WriteLegacy(w io.Writer, m Manifest) error {
	names := make([]string, 0, len(m.Entries))

	for name := range m.Entries {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		entry := m.Entries[name]

		_, err := fmt.Fprintln(w, entry.MD5, entry.Name)
		if err != nil {
			return fmt.Errorf("failed to write entry %s: %w", name, err)
		}
	}

	return nil
}

func ReadLegacy(r io.Reader) (Manifest, error) {
	m := Manifest{
		Entries: make(map[string]Entry),
	}

	scanner := bufio.NewScanner(r)
	lineCount := 0
	for scanner.Scan() {
		lineCount++

		line := scanner.Text()
		if line == "" {
			continue
		}

		// expected string: MD5 FILENAME
		sum, name, found := strings.Cut(line, " ")
		if !found || sum == "" || name == "" {
			return Manifest{}, fmt.Errorf("invalid manifest entry at line %d", lineCount)
		}

		// this should never happens
		if _, exists := m.Entries[name]; exists {
			return Manifest{}, fmt.Errorf("duplicate entry for %s at line %d", name, lineCount)
		}

		m.Entries[name] = Entry{
			Name: name,
			MD5:  sum,
		}
	}

	if err := scanner.Err(); err != nil {
		return m, fmt.Errorf("failed to read legacy manifest: %w", err)
	}

	return m, nil
}
