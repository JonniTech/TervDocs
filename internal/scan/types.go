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
	PrimaryLanguage   string
	MainLanguages     []string
	Frameworks        []string
	Dependencies      []string
	PackageManager    string
	EntryPoints       []string
	Scripts           map[string]string
	UsesDocker        bool
	CIConfigs         []string
	ConfigFiles       []string
	EnvFiles          []string
	ReadmeExists      bool
	TestSetup         []string
	MonorepoHints     []string
	RouteHints        []string
	TopFolders        []string
	FolderFileCounts  map[string]int
	PurposeHints      []string
	ArchitectureHints []string
	ImportantFiles    []FileSummary
	Warnings          []string
	ExcludedByDefault []string
}
