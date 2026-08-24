package httpapi

import (
	"net/http"

	"task225-timberjoint/internal/service"
)

// createBatch POST /api/batches
func (s *Server) createBatch(w http.ResponseWriter, r *http.Request) {
	var in service.BatchCreateInput
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	b, err := s.app.Batches.Create(in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

// listBatches GET /api/batches
func (s *Server) listBatches(w http.ResponseWriter, r *http.Request) {
	batches, err := s.app.Batches.List()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, batches)
}

// getBatch GET /api/batches/{id}
func (s *Server) getBatch(w http.ResponseWriter, r *http.Request) {
	b, err := s.app.Batches.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// advanceBatch POST /api/batches/{id}/advance
func (s *Server) advanceBatch(w http.ResponseWriter, r *http.Request) {
	b, err := s.app.Batches.Advance(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// archiveBatch POST /api/batches/{id}/archive
func (s *Server) archiveBatch(w http.ResponseWriter, r *http.Request) {
	b, err := s.app.Batches.Archive(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}
