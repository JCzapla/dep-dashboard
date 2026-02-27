package http

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/JCzapla/dep-dashboard/internal/domain"
	"github.com/JCzapla/dep-dashboard/internal/port/inbound"
)


type indexData struct {
	Package *domain.Package
	Filter string
	MinScore string
	Error string
}

type Handler struct {
	service inbound.DependencyService
	tmpl *template.Template
	config Config
}

func NewHandler(service inbound.DependencyService, tmpl *template.Template, cfg Config) *Handler {
	return &Handler{service: service, tmpl: tmpl, config: cfg}
}

func (h *Handler) PutDeps(w http.ResponseWriter, r *http.Request) {
	req := PostDepsRequest{
		Name: h.config.DefaultPackage.Name,
	}
	name := r.PathValue("name")
	hasBody := r.ContentLength != 0
	hasPath := name != ""
	if hasBody && hasPath {
		writeJSON(w, http.StatusBadRequest, "Used both: URL Path Param and Body. Use one")
		return
	}
	if hasBody {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if req.Name == "" {
			req.Name = h.config.DefaultPackage.Name
		}
	} else if hasPath {
		req.Name = name
	}


	pkg, err := h.service.StoreDependencies(r.Context(), req.Name)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, fmt.Errorf("%w", err))
		return
	}
	writeJSON(w, http.StatusCreated, toResponse(pkg))
}



func (h *Handler) GetDeps(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var filters []domain.Filter
	if param := q.Get("name"); param != "" {
		filters = append(filters, domain.Filter{Column: "name", Operator: domain.FilterEq, Value: param})
	}
	if param := q.Get("minScore"); param != "" {
		filters = append(filters, domain.Filter{Column: "score", Operator: domain.FilterGte, Value: param})
	}

	data := indexData{Filter: q.Get("name"), MinScore: q.Get("minScore")}
	pkg, err := h.service.GetDependencies(r.Context(), filters)
	if err != nil {
		data.Error = err.Error()
	} else {
		data.Package = pkg
	}

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "text/html")
		h.tmpl.ExecuteTemplate(w, "index.html", data)
		return
	}

	if data.Error != "" {
		writeJSON(w, http.StatusBadRequest, data.Error)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(pkg))
}

func (h *Handler) DeleteDeps(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	err := h.service.DeleteDependenciesByName(r.Context(), name)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}


func toResponse(pkg *domain.Package) DepsResponse {
	nodes := make([]DependencyNode, len(pkg.Dependencies))
	for i, n := range pkg.Dependencies {
		nodes[i] = DependencyNode{
			Name: n.Name,
			Version: n.Version,
			Relation: n.Relation,
			Score: n.Score,
		}
	}
	return DepsResponse{
		ID: pkg.ID,
		Name: pkg.PackageRef.Name,
		Version: pkg.PackageRef.Version,
		Dependencies: nodes,
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}