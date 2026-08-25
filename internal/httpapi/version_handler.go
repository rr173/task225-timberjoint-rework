package httpapi

import (
	"net/http"

	"task225-timberjoint/internal/model"
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
	// 串行化同节点的并发发布：在分配版本号与替代旧冻结版本的关键区内
	// 不可交错，否则会因抢占同一 version_no 丢失其中一个发布。
	var v *model.NodeVersion
	err := s.app.WithLock(func() error {
		var err error
		v, err = s.app.Versions.Publish(r.PathValue("id"), r.PathValue("nodeNo"), body.Reason)
		return err
	})
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
