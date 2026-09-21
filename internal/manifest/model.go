package manifest

import "maps"

type Entry struct {
	Name string
	MD5  string
}

type Manifest struct {
	Entries map[string]Entry
}

const LegacyMetadataFile = ".dir_md5.txt"

func (m *Manifest) Merge(other Manifest) {
	if m.Entries == nil {
		m.Entries = make(map[string]Entry, len(other.Entries))
	}
	maps.Copy(m.Entries, other.Entries)
}
