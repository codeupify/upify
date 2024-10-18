package template

import (
	_ "embed"
)

//go:embed python_main.template
var PythonMainTemplate string

//go:embed node_main.template
var NodeMainTemplate string
