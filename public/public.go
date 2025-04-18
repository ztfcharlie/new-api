package public

import "embed"

//go:embed webHtml/*.html
var TemplatesFs embed.FS
