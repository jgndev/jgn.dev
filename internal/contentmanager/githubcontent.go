package contentmanager

// githubContent represents a content item in a GitHub repository, which may be a file or a directory.
type githubContent struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
	Size int    `json:"size"`
}
