package httpapi

import (
	"net/http"
	"time"
)

// addEvidence POST /api/joints/{id}/evidences
func (s *Server) addEvidence(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Filename string `json:"filename"`
		Caption  string `json:"caption"`
		TakenAt  string `json:"taken_at"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, err)
		return
	}
	takenAt, err := time.Parse(time.RFC3339, body.TakenAt)
	if err != nil {
		writeError(w, err)
		return
	}
	e, err := s.app.Evidences.Add(r.PathValue("id"), body.Filename, body.Caption, takenAt)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

// listEvidences GET /api/joints/{id}/evidences
func (s *Server) listEvidences(w http.ResponseWriter, r *http.Request) {
	evidences, err := s.app.Evidences.ListByJoint(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, evidences)
}
