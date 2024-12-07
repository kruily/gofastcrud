package templates

import (
	"embed"
	"html/template"
	"io"
)

//go:embed home/home.html swagger-ui/*
var Templates embed.FS

// SwaggerUITemplate 渲染Swagger UI模板
func SwaggerUITemplate(w io.Writer, data interface{}) error {
	tmpl, err := template.ParseFS(Templates, "swagger-ui/index.html")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}
