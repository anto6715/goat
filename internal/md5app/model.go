package md5app

type hashResult struct {
	path string
	sum  string
	err  error
}

type hashJob struct {
	path string
}
