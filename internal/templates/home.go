package templates

import (
	"html/template"
	"net/http"
	"runtime"
)

// HomeData 主页数据
type HomeData struct {
	GoVersion   string
	APIVersions []string
}

// HomeHandler 返回主页 HTML
func HomeHandler(apiVersions []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(Templates, "home/home.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := HomeData{
			GoVersion:   runtime.Version(),
			APIVersions: apiVersions,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
