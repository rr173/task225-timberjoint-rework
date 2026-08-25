package httpapi

import (
	"net/http"
)

// publishVersion POST /api/batches/{id}/nodes/{nodeNo}/versions
func (s *Server) publishVersion(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Reason string `json:"reason"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, err)
		return
	}
	v, err := s.app.Versions.Publish(r.PathValue("id"), r.PathValue("nodeNo"), body.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

// listNodeVersions GET /api/batches/{id}/nodes/{nodeNo}/versions
func (s *Server) listNodeVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := s.app.Versions.ListByNode(r.PathValue("id"), r.PathValue("nodeNo"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, versions)
}

// getVersion GET /api/versions/{id}
func (s *Server) getVersion(w http.ResponseWriter, r *http.Request) {
	v, err := s.app.Versions.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}
