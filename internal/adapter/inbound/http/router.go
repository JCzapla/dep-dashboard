package http

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/JCzapla/dep-dashboard/internal/domain"
	"github.com/JCzapla/dep-dashboard/internal/port/inbound"
)

type Config struct {
	DefaultPackage domain.PackageRef
}

//go:embed templates
var templateFiles embed.FS

var templateFuncs = template.FuncMap{
	"scoreBarWidth": func(score *float64) string {
		if score == nil {
			return "0"
		}
		return fmt.Sprintf("%.1f", *score*10)
	},
	"scoreBarColor": func(score *float64) string {
		if score == nil {
			return "score-nil"
		}
		switch {
		case *score >= 7.5:
			return "score-green"
		case *score >= 4.0:
			return "score-yellow"
		default:	
			return "score-red"
		}
	},
}

func NewRouter(service inbound.DependencyService, cfg Config) *http.ServeMux {
	tmpl := template.Must(template.New("").Funcs(templateFuncs).ParseFS(templateFiles, "templates/*.html"))
	h := NewHandler(service, tmpl, cfg)
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.GetDeps)
	mux.HandleFunc("/deps", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut:
			h.PutDeps(w, r)
		case http.MethodGet:
			h.GetDeps(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"Error": "Method not allowed"}`))
		}
	})
	mux.HandleFunc("/deps/{name}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			h.PutDeps(w, r)
		case http.MethodGet:
			h.GetDeps(w, r)
		case http.MethodDelete:
			h.DeleteDeps(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"Error": "Method not allowed"}`))
		}
	})
	return mux
}