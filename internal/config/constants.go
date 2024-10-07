package config

type Language string

const (
	Python     Language = "python"
	JavaScript Language = "javascript"
	TypeScript Language = "typescript"
)

type Framework string

const (
	Flask   Framework = "flask"
	Express Framework = "express"
)

type PackageManager string

const (
	Pip  PackageManager = "pip"
	Npm  PackageManager = "npm"
	Yarn PackageManager = "yarn"
)

type ModuleSystem string

const (
	CommonJS ModuleSystem = "commonjs"
	ES6      ModuleSystem = "es6"
)
