package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task225-timberjoint/internal/service"
	"task225-timberjoint/internal/store"
)

func TestHandlerServesReviewPageAndHealthAPI(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	app, err := service.New(db)
	if err != nil {
		t.Fatalf("create app: %v", err)
	}
	handler := New(app).Handler()

	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "古建筑木构节点测绘复核") {
		t.Fatalf("review page unavailable: status=%d body=%q", page.Code, page.Body.String())
	}

	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"status":"ok"`) {
		t.Fatalf("health API unavailable: status=%d body=%q", health.Code, health.Body.String())
	}
}
