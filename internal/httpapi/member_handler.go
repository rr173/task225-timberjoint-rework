package httpapi

import (
	"net/http"

	"task225-timberjoint/internal/service"
)

// addMember POST /api/batches/{id}/members
func (s *Server) addMember(w http.ResponseWriter, r *http.Request) {
	var in service.MemberAddInput
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	m, err := s.app.Members.Add(r.PathValue("id"), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

// listMembers GET /api/batches/{id}/members
func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	members, err := s.app.Members.ListByBatch(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, members)
}

// getMember GET /api/members/{id}
func (s *Server) getMember(w http.ResponseWriter, r *http.Request) {
	m, err := s.app.Members.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// markMemberStatus PATCH /api/members/{id}/status
func (s *Server) markMemberStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, err)
		return
	}
	m, err := s.app.Members.MarkStatus(r.PathValue("id"), body.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}
