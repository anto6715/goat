package manifest

type Entry struct {
	Name string
	MD5  string
}

type Manifest struct {
	Entries map[string]Entry
}

const LegacyMetadataFile = ".dir_md5.txt"
