package scan

type FileSummary struct {
	Path    string
	Size    int64
	Snippet string
}

type RepoSummary struct {
	Root              string
	ProjectName       string
	FilesScanned      int
	BytesScanned      int64
	MainLanguages     []string
	Frameworks        []string
	PackageManager    string
	EntryPoints       []string
	Scripts           map[string]string
	UsesDocker        bool
	CIConfigs         []string
	EnvFiles          []string
	ReadmeExists      bool
	TestSetup         []string
	MonorepoHints     []string
	RouteHints        []string
	TopFolders        []string
	ImportantFiles    []FileSummary
	Warnings          []string
	ExcludedByDefault []string
}
