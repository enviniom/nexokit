package roles

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

func jsonRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	r := httptest.NewRequest(method, path, &buf)
	r.Header.Set("Content-Type", "application/json")
	return r
}

// fakeService is a test double for the service.
type fakeService struct {
	roles      []RoleResponse
	role       *RoleResponse
	total      int64
	err        error
	created    *RoleResponse
	updated    *RoleResponse
	deletedPID string
}

func (f *fakeService) List(page, perPage int) ([]RoleResponse, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.roles, f.total, nil
}

func (f *fakeService) GetByPublicID(publicID string) (*RoleResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.role, nil
}

func (f *fakeService) Create(req CreateRoleRequest) (*RoleResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.created, nil
}

func (f *fakeService) Update(publicID string, req UpdateRoleRequest) (*RoleResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.updated != nil {
		return f.updated, nil
	}
	return f.role, nil
}

func (f *fakeService) Delete(publicID string) error {
	f.deletedPID = publicID
	return f.err
}

func setupHandler(svc Service) (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(svc)
	return gin.New(), h
}

func TestHandler_List(t *testing.T) {
	t.Run("returns paginated roles list", func(t *testing.T) {
		svc := &fakeService{
			roles: []RoleResponse{
				{PublicID: "role1", Name: "admin", IsSystem: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{PublicID: "role2", Name: "user", IsSystem: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
			total: 2,
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/roles?page=1&per_page=10", nil)
		h.List(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.APIResponse[[]RoleResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Error("expected success response")
		}
		metaMap, ok := resp.Meta.(map[string]any)
		if !ok {
			t.Fatal("expected meta to be a map")
		}
		pgRaw, ok := metaMap["pagination"]
		if !ok {
			t.Fatal("expected pagination in meta")
		}
		pgBytes, _ := json.Marshal(pgRaw)
		var pg response.PaginationMeta
		if err := json.Unmarshal(pgBytes, &pg); err != nil {
			t.Fatalf("failed to unmarshal pagination: %v", err)
		}
		if pg.Total != 2 {
			t.Errorf("expected total 2, got %d", pg.Total)
		}
		if len(resp.Data) != 2 {
			t.Errorf("expected 2 roles, got %d", len(resp.Data))
		}
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := &fakeService{err: errors.New("boom")}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/roles", nil)
		h.List(c)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_GetByPublicID(t *testing.T) {
	t.Run("returns role when found", func(t *testing.T) {
		svc := &fakeService{
			role: &RoleResponse{PublicID: "role1", Name: "admin", IsSystem: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "role1"}}
		h.GetByPublicID(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.APIResponse[RoleResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Error("expected success response")
		}
		if resp.Data.PublicID != "role1" {
			t.Errorf("expected public_id 'role1', got %s", resp.Data.PublicID)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrNotFound}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		h.GetByPublicID(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}

func TestHandler_Create(t *testing.T) {
	t.Run("creates role successfully", func(t *testing.T) {
		svc := &fakeService{
			created: &RoleResponse{PublicID: "role3", Name: "editor", Slug: "editor"},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/roles", CreateRoleRequest{Name: "editor", Slug: "editor"})
		h.Create(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var resp response.APIResponse[RoleResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Error("expected success response")
		}
		if resp.Data.PublicID != "role3" {
			t.Errorf("expected public_id 'role3', got %s", resp.Data.PublicID)
		}
	})

	t.Run("returns conflict on duplicate name", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrConflict}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/roles", CreateRoleRequest{Name: "editor", Slug: "editor"})
		h.Create(c)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("returns validation error on invalid fields", func(t *testing.T) {
		svc := &fakeService{created: &RoleResponse{PublicID: "role3", Name: "editor", Slug: "editor"}}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/roles", CreateRoleRequest{Name: "", Slug: ""})
		h.Create(c)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", w.Code)
		}

		var resp response.APIResponse[any]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Success {
			t.Error("expected error response")
		}
		errsMap, ok := resp.Errors.(map[string]any)
		if !ok {
			t.Fatal("expected errors to be a map")
		}
		if _, ok := errsMap["name"]; !ok {
			t.Error("expected validation error for name")
		}
		if _, ok := errsMap["slug"]; !ok {
			t.Error("expected validation error for slug")
		}
	})
}

func TestHandler_Update(t *testing.T) {
	t.Run("updates role successfully", func(t *testing.T) {
		svc := &fakeService{
			updated: &RoleResponse{PublicID: "role1", Name: "senior-editor", Slug: "senior-editor"},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "role1"}}
		c.Request = jsonRequest(http.MethodPut, "/roles/role1", UpdateRoleRequest{Name: "senior-editor", Slug: "senior-editor"})
		h.Update(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.APIResponse[RoleResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Data.Name != "senior-editor" {
			t.Errorf("expected name 'senior-editor', got %s", resp.Data.Name)
		}
	})

	t.Run("returns forbidden for system role", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrForbidden}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "role1"}}
		c.Request = jsonRequest(http.MethodPut, "/roles/role1", UpdateRoleRequest{Name: "super-root", Slug: "super-root"})
		h.Update(c)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("returns validation error on invalid fields", func(t *testing.T) {
		svc := &fakeService{updated: &RoleResponse{PublicID: "role1", Name: "editor", Slug: "editor"}}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "role1"}}
		c.Request = jsonRequest(http.MethodPut, "/roles/role1", UpdateRoleRequest{Name: "", Slug: ""})
		h.Update(c)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", w.Code)
		}

		var resp response.APIResponse[any]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Success {
			t.Error("expected error response")
		}
		errsMap, ok := resp.Errors.(map[string]any)
		if !ok {
			t.Fatal("expected errors to be a map")
		}
		if _, ok := errsMap["name"]; !ok {
			t.Error("expected validation error for name")
		}
		if _, ok := errsMap["slug"]; !ok {
			t.Error("expected validation error for slug")
		}
	})
}

func TestHandler_Delete(t *testing.T) {
	t.Run("deletes role successfully", func(t *testing.T) {
		svc := &fakeService{}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "role1"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/roles/role1", nil)
		h.Delete(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if svc.deletedPID != "role1" {
			t.Errorf("expected deleted public_id 'role1', got %s", svc.deletedPID)
		}
	})

	t.Run("returns forbidden for system role", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrForbidden}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "role1"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/roles/role1", nil)
		h.Delete(c)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", w.Code)
		}
	})
}
