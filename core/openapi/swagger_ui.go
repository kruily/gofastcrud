package openapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-openapi/spec"
	"github.com/kruily/gofastcrud/core/templates"
)

type VersionInfo struct {
	Version string
	Path    string
}

// SwaggerUIHandler 处理Swagger UI请求
func SwaggerUIHandler(w http.ResponseWriter, r *http.Request, versions []string, currentVersion string, docs interface{}) {
	path := r.URL.Path
	path = strings.TrimPrefix(path, fmt.Sprintf("/api/%s/swagger/", currentVersion))

	// 处理doc.json请求
	if path == "doc.json" {
		w.Header().Set("Content-Type", "application/json")
		if allDocs, ok := docs.(map[string]*spec.Swagger); ok {
			if doc, exists := allDocs[currentVersion]; exists {
				json.NewEncoder(w).Encode(doc)
			}
		}
		if allDocs, ok := docs.(map[string]*openapi3.T); ok {
			if doc, exists := allDocs[currentVersion]; exists {
				json.NewEncoder(w).Encode(doc)
			}
		}
		return
	}

	// 处理index.html
	if path == "index.html" || path == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// 构建版本信息
		versionInfos := make([]VersionInfo, 0, len(versions))
		for _, v := range versions {
			versionInfos = append(versionInfos, VersionInfo{
				Version: v,
				Path:    fmt.Sprintf("/api/%s/swagger/", v),
			})
		}

		data := map[string]interface{}{
			"URL":            fmt.Sprintf("/api/%s/swagger/doc.json", currentVersion),
			"Versions":       versionInfos,
			"CurrentVersion": currentVersion,
		}

		if err := templates.SwaggerUITemplate(w, data); err != nil {
			log.Printf("Template error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 处理其他静态资源
	content, err := templates.Templates.ReadFile("swagger-ui/" + path)
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
