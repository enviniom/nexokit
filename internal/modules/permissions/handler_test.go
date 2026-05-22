package permissions

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

// fakePermissionService is a test double for the Service interface.
type fakePermissionService struct {
	groups      []PermissionGroupResponse
	permissions []PermissionResponse
	total       int64
	single      *PermissionResponse
	err         error
}

func (f *fakePermissionService) ListGrouped() ([]PermissionGroupResponse, error) {
	return f.groups, f.err
}

func (f *fakePermissionService) List(params query.ListParams) ([]PermissionResponse, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.permissions, f.total, nil
}

func (f *fakePermissionService) GetByPublicID(publicID string) (*PermissionResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.single, nil
}

func (f *fakePermissionService) Create(req CreatePermissionRequest) (*PermissionResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.single, nil
}

func (f *fakePermissionService) Update(publicID string, req UpdatePermissionRequest) (*PermissionResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.single, nil
}

func (f *fakePermissionService) Delete(publicID string) error {
	return f.err
}

func (f *fakePermissionService) Resolve(publicID string) ([]string, error) {
	return nil, f.err
}

func setupPermissionHandler(svc Service) *Handler {
	gin.SetMode(gin.TestMode)
	return NewHandler(svc)
}

func jsonPermissionRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	r := httptest.NewRequest(method, path, &buf)
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestPermissionHandler_ListPaginated(t *testing.T) {
	t.Run("returns paginated permissions with filters metadata", func(t *testing.T) {
		svc := &fakePermissionService{
			permissions: []PermissionResponse{
				{PublicID: "p1", Slug: "users.index", Module: "users", Action: ActionIndex},
				{PublicID: "p2", Slug: "users.create", Module: "users", Action: ActionCreate},
			},
			total: 2,
		}
		h := setupPermissionHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/permissions?page=1&per_page=10", nil)
		h.ListPaginated(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.PaginatedResponse[[]PermissionResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Error("expected success response")
		}
		if len(resp.Data) != 2 {
			t.Errorf("expected 2 permissions, got %d", len(resp.Data))
		}
		metaMap, ok := resp.Meta.(map[string]any)
		if !ok {
			t.Fatal("expected meta to be a map")
		}
		if _, ok := metaMap["pagination"]; !ok {
			t.Error("expected pagination in meta")
		}
		if _, ok := metaMap["filters"]; !ok {
			t.Error("expected filters in meta")
		}
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		svc := &fakePermissionService{err: apperror.ErrInternal}
		h := setupPermissionHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/permissions", nil)
		h.ListPaginated(c)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})
}

func TestPermissionHandler_GetByPublicID(t *testing.T) {
	t.Run("returns permission when found", func(t *testing.T) {
		svc := &fakePermissionService{
			single: &PermissionResponse{PublicID: "p1", Slug: "users.index", Module: "users", Action: ActionIndex},
		}
		h := setupPermissionHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "p1"}}
		h.GetByPublicID(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		var resp response.APIResponse[PermissionResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Data.PublicID != "p1" {
			t.Errorf("expected public_id p1, got %s", resp.Data.PublicID)
		}
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		svc := &fakePermissionService{err: apperror.ErrNotFound}
		h := setupPermissionHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		h.GetByPublicID(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}

func TestPermissionHandler_Create(t *testing.T) {
	t.Run("creates permission successfully", func(t *testing.T) {
		svc := &fakePermissionService{
			single: &PermissionResponse{PublicID: "p1", Slug: "users.export", Module: "users", Action: "export"},
		}
		h := setupPermissionHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonPermissionRequest(http.MethodPost, "/permissions", CreatePermissionRequest{
			Slug: "users.export", Name: "Export users", Module: "users", Action: "export",
		})
		h.Create(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}
	})

	t.Run("returns 409 on duplicate slug", func(t *testing.T) {
		svc := &fakePermissionService{err: apperror.ErrConflict}
		h := setupPermissionHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonPermissionRequest(http.MethodPost, "/permissions", CreatePermissionRequest{
			Slug: "users.index", Name: "Index users", Module: "users", Action: ActionIndex,
		})
		h.Create(c)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("returns 403 for forbidden operation", func(t *testing.T) {
		svc := &fakePermissionService{err: apperror.ErrForbidden}
		h := setupPermissionHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonPermissionRequest(http.MethodPost, "/permissions", CreatePermissionRequest{
			Slug: "users.index", Name: "Index users", Module: "users", Action: ActionIndex,
		})
		h.Create(c)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", w.Code)
		}
	})
}

func TestPermissionHandler_Delete(t *testing.T) {
	t.Run("deletes permission successfully", func(t *testing.T) {
		svc := &fakePermissionService{}
		h := setupPermissionHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "p1"}}
		h.Delete(c)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d", w.Code)
		}
		if w.Body.Len() != 0 {
			t.Fatalf("expected empty body, got %q", w.Body.String())
		}
	})

	t.Run("returns 403 when system permission", func(t *testing.T) {
		svc := &fakePermissionService{err: apperror.ErrForbidden}
		h := setupPermissionHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "system-p"}}
		h.Delete(c)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		svc := &fakePermissionService{err: apperror.ErrNotFound}
		h := setupPermissionHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		h.Delete(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}
