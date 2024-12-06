package swagger

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed swagger-ui/*
var swaggerUI embed.FS

// SwaggerUIHandler 处理Swagger UI请求
func SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/swagger/" || path == "/swagger" {
		path = "/swagger/index.html"
	}

	// 移除前缀
	path = path[len("/swagger/"):]

	// 处理index.html
	if path == "index.html" {
		tmpl, err := template.ParseFS(swaggerUI, "swagger-ui/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, map[string]interface{}{
			"URL": "/swagger/doc.json",
		})
		return
	}

	// 处理其他静态资源
	content, err := swaggerUI.ReadFile("swagger-ui/" + path)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// 设置正确的Content-Type
	switch {
	case path[len(path)-3:] == ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case path[len(path)-4:] == ".css":
		w.Header().Set("Content-Type", "text/css")
	case path[len(path)-5:] == ".html":
		w.Header().Set("Content-Type", "text/html")
	}

	w.Write(content)
}
