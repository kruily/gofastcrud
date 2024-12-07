package templates

import "embed"

//go:embed home/home.html swagger-ui/index.html
var Templates embed.FS
