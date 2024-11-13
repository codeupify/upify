package python

import (
	_ "embed"
)

//go:embed templates/python_main.tmpl
var PythonMainTemplate string
