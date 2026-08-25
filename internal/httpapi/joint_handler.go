package httpapi

import "net/http"

// createJoint POST /api/batches/{id}/joints
func (s *Server) createJoint(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NodeNo    string   `json:"node_no"`
		PointIDs  []string `json:"point_ids"`
		MemberIDs []string `json:"member_ids"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, err)
		return
	}
	j, err := s.app.Joints.Create(r.PathValue("id"), body.NodeNo, body.PointIDs, body.MemberIDs)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, j)
}

// listJoints GET /api/batches/{id}/joints
func (s *Server) listJoints(w http.ResponseWriter, r *http.Request) {
	joints, err := s.app.Joints.ListByBatch(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, joints)
}

// getJoint GET /api/joints/{id}
func (s *Server) getJoint(w http.ResponseWriter, r *http.Request) {
	j, err := s.app.Joints.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, j)
}

// checkJoint POST /api/joints/{id}/check
func (s *Server) checkJoint(w http.ResponseWriter, r *http.Request) {
	rep, err := s.app.Joints.Check(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

// confirmJoint POST /api/joints/{id}/confirm
func (s *Server) confirmJoint(w http.ResponseWriter, r *http.Request) {
	var j any
	err := s.app.WithLock(func() error {
		var err error
		j, err = s.app.Joints.Confirm(r.PathValue("id"))
		return err
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, j)
}

// rejectJoint POST /api/joints/{id}/reject
func (s *Server) rejectJoint(w http.ResponseWriter, r *http.Request) {
	j, err := s.app.Joints.Reject(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, j)
}
