package node

import (
	_ "embed"
)

//go:embed templates/node_main.tmpl
var NodeMainTemplate string
