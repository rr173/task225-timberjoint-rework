// Package httpapi 提供 HTTP 层：路由注册、请求解析与 JSON 响应。
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"task225-timberjoint/internal/model"
	"task225-timberjoint/internal/service"
)

// Server HTTP 服务器。
type Server struct {
	app *service.App
}

// New 构造 HTTP 服务器。
func New(app *service.App) *Server { return &Server{app: app} }

// Handler 注册全部路由并返回处理器。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// 批次生命周期。
	mux.HandleFunc("POST /api/batches", s.createBatch)
	mux.HandleFunc("GET /api/batches", s.listBatches)
	mux.HandleFunc("GET /api/batches/{id}", s.getBatch)
	mux.HandleFunc("POST /api/batches/{id}/advance", s.advanceBatch)
	mux.HandleFunc("POST /api/batches/{id}/archive", s.archiveBatch)

	// 测点。
	mux.HandleFunc("POST /api/batches/{id}/points", s.addPoint)
	mux.HandleFunc("GET /api/batches/{id}/points", s.listPoints)
	mux.HandleFunc("GET /api/points/{id}", s.getPoint)
	mux.HandleFunc("PATCH /api/points/{id}/status", s.markPointStatus)

	// 构件。
	mux.HandleFunc("POST /api/batches/{id}/members", s.addMember)
	mux.HandleFunc("GET /api/batches/{id}/members", s.listMembers)
	mux.HandleFunc("GET /api/members/{id}", s.getMember)
	mux.HandleFunc("PATCH /api/members/{id}/status", s.markMemberStatus)

	// 节点关系与几何复核。
	mux.HandleFunc("POST /api/batches/{id}/joints", s.createJoint)
	mux.HandleFunc("GET /api/batches/{id}/joints", s.listJoints)
	mux.HandleFunc("GET /api/joints/{id}", s.getJoint)
	mux.HandleFunc("POST /api/joints/{id}/check", s.checkJoint)
	mux.HandleFunc("POST /api/joints/{id}/confirm", s.confirmJoint)
	mux.HandleFunc("POST /api/joints/{id}/reject", s.rejectJoint)

	// 现场照片证据。
	mux.HandleFunc("POST /api/joints/{id}/evidences", s.addEvidence)
	mux.HandleFunc("GET /api/joints/{id}/evidences", s.listEvidences)

	// 节点版本发布。
	mux.HandleFunc("POST /api/batches/{id}/nodes/{nodeNo}/versions", s.publishVersion)
	mux.HandleFunc("GET /api/batches/{id}/nodes/{nodeNo}/versions", s.listNodeVersions)
	mux.HandleFunc("GET /api/versions/{id}", s.getVersion)

	// 统计与健康。
	mux.HandleFunc("GET /api/stats", s.stats)
	mux.HandleFunc("GET /api/health", s.health)

	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "task225-timberjoint"})
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	st, err := s.app.Stats()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// writeJSON 输出 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError 把领域错误映射为 HTTP 状态码并输出。
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrInvalidInput):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrInvalidState), errors.Is(err, model.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, model.ErrSealed):
		status = http.StatusConflict
	case errors.Is(err, model.ErrDuplicate):
		status = http.StatusConflict
	case errors.Is(err, model.ErrInsufficientData), errors.Is(err, model.ErrGeometry):
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// decode 解析 JSON 请求体。
func decode(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
