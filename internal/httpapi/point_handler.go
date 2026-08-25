package httpapi

import (
	"net/http"

	"task225-timberjoint/internal/service"
)

// addPoint POST /api/batches/{id}/points
func (s *Server) addPoint(w http.ResponseWriter, r *http.Request) {
	var in service.PointAddInput
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	p, err := s.app.Points.Add(r.PathValue("id"), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

// listPoints GET /api/batches/{id}/points
func (s *Server) listPoints(w http.ResponseWriter, r *http.Request) {
	points, err := s.app.Points.ListByBatch(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, points)
}

// getPoint GET /api/points/{id}
func (s *Server) getPoint(w http.ResponseWriter, r *http.Request) {
	p, err := s.app.Points.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// markPointStatus PATCH /api/points/{id}/status
func (s *Server) markPointStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, err)
		return
	}
	p, err := s.app.Points.MarkStatus(r.PathValue("id"), body.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}
