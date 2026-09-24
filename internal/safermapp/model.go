package safermapp

type fileKey struct {
	name string
	md5  string
}

type fileRecord struct {
	path        string
	manifestDir string
}

type fileIndex map[fileKey][]fileRecord

type candidate struct {
	references []fileRecord
	target     fileRecord
	key        fileKey
}
