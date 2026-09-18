package manifest

import (
	"fmt"
	"io"
	"sort"
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
