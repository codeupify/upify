package lang

type Language string

const (
	Python     Language = "python"
	JavaScript Language = "javascript"
	TypeScript Language = "typescript"
)

type PackageManager string

const (
	Pip  PackageManager = "pip"
	Npm  PackageManager = "npm"
	Yarn PackageManager = "yarn"
)
