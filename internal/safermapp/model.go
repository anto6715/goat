package safermapp

type HashFiles struct {
	data   map[string][]string
	nfiles int
}

// Constructor
func NewHashFiles() HashFiles {
	return HashFiles{make(map[string][]string), 0}
}

// Adds a path to the hash files
func (h *HashFiles) AddPath(hash string, path string) {
	h.data[hash] = append(h.data[hash], path)
	h.nfiles++
}

// Merges another HashFiles into this one.
// If a hash exists in both, the paths are merged.
func (h *HashFiles) Merge(other HashFiles) {
	for hash, paths := range other.data {
		h.data[hash] = append(h.data[hash], paths...)
	}
	h.nfiles += other.nfiles
}

// Returns a slice of all the hashes in the HashFiles.
func (h *HashFiles) Hashes() []string {
	var hashes []string
	for hash := range h.data {
		hashes = append(hashes, hash)
	}
	return hashes
}

// Returns the number of files in the HashFiles.
func (h *HashFiles) NFiles() int {
	return h.nfiles
}

// Returns true if the HashFiles contains the given hash.
func (h *HashFiles) Contains(hash string) bool {
	_, ok := h.data[hash]
	return ok
}

// Returns the paths associated with the given hash.
func (h *HashFiles) GetPaths(hash string) []string {
	return h.data[hash]
}
