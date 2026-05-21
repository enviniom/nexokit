package users

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
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
	users             []UserResponse
	user              *UserResponse
	total             int64
	err               error
	created           *UserResponse
	updated           *UserResponse
	deletedPID        string
	changePasswordErr error
	toggled           *UserResponse
	updateActor       string
	passwordActor     string
	tenant            tenant.TenantContext
	createReq         CreateUserRequest
	listParams        query.ListParams
}

func (f *fakeService) List(tc tenant.TenantContext, params query.ListParams) ([]UserResponse, int64, error) {
	f.tenant = tc
	f.listParams = params
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.users, f.total, nil
}

func (f *fakeService) GetByPublicID(tc tenant.TenantContext, publicID string) (*UserResponse, error) {
	f.tenant = tc
	if f.err != nil {
		return nil, f.err
	}
	return f.user, nil
}

func (f *fakeService) Create(tc tenant.TenantContext, req CreateUserRequest) (*UserResponse, error) {
	f.tenant = tc
	f.createReq = req
	if f.err != nil {
		return nil, f.err
	}
	return f.created, nil
}

func (f *fakeService) Update(tc tenant.TenantContext, publicID string, actorPublicID string, req UpdateUserRequest) (*UserResponse, error) {
	f.tenant = tc
	f.updateActor = actorPublicID
	if f.err != nil {
		return nil, f.err
	}
	if f.updated != nil {
		return f.updated, nil
	}
	return f.user, nil
}

func (f *fakeService) Delete(tc tenant.TenantContext, publicID string) error {
	f.tenant = tc
	f.deletedPID = publicID
	return f.err
}

func (f *fakeService) ChangePassword(tc tenant.TenantContext, publicID string, actorPublicID string, req ChangePasswordRequest) error {
	f.tenant = tc
	f.passwordActor = actorPublicID
	return f.changePasswordErr
}

func (f *fakeService) ToggleStatus(tc tenant.TenantContext, publicID string, req UpdateStatusRequest) (*UserResponse, error) {
	f.tenant = tc
	if f.err != nil {
		return nil, f.err
	}
	return f.toggled, nil
}

func setTenant(c *gin.Context) {
	tenant.SetGin(c, tenant.NewScoped(7, "acme"))
}

func setupHandler(svc Service) (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(svc, nil)
	return gin.New(), h
}

func TestHandler_List(t *testing.T) {
	t.Run("returns paginated users list", func(t *testing.T) {
		svc := &fakeService{
			users: []UserResponse{
				{PublicID: "user1", Name: "Alice", Email: "alice@example.com", IsActive: true, RoleID: 1, RoleName: "admin", CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{PublicID: "user2", Name: "Bob", Email: "bob@example.com", IsActive: true, RoleID: 2, RoleName: "user", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
			total: 2,
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users?page=1&per_page=10", nil)
		setTenant(c)
		h.List(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.APIResponse[[]UserResponse]
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
			t.Errorf("expected 2 users, got %d", len(resp.Data))
		}
	})

	t.Run("passes filters search and sorting params to service and returns filter metadata", func(t *testing.T) {
		svc := &fakeService{
			users: []UserResponse{{PublicID: "user1", Name: "Alice", Email: "alice@example.com", IsActive: true, RoleID: 2, RoleName: "admin", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
			total: 1,
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users?page=2&per_page=5&status=active&created_from=2025-01-01&created_to=2025-12-31&sort=name&order=asc&search=ali", nil)
		setTenant(c)
		h.List(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
		}
		if svc.listParams.Pagination.Page != 2 || svc.listParams.Pagination.PerPage != 5 {
			t.Fatalf("expected parsed pagination, got %#v", svc.listParams.Pagination)
		}
		if svc.listParams.Filters.Status != "active" || svc.listParams.Filters.CreatedFrom == nil || svc.listParams.Filters.CreatedTo == nil {
			t.Fatalf("expected status and date filters, got %#v", svc.listParams.Filters)
		}
		if svc.listParams.Sort.Sort != "name" || svc.listParams.Sort.Order != "asc" || svc.listParams.Search.Query != "ali" {
			t.Fatalf("expected sort/search params, got sort=%#v search=%#v", svc.listParams.Sort, svc.listParams.Search)
		}

		var resp response.APIResponse[[]UserResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		metaMap := resp.Meta.(map[string]any)
		filters := metaMap["filters"].(map[string]any)
		if filters["status"] != "active" || filters["sort"] != "name" || filters["order"] != "asc" || filters["search"] != "ali" {
			t.Fatalf("expected filter metadata to reflect query, got %#v", filters)
		}
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := &fakeService{err: errors.New("boom")}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users", nil)
		setTenant(c)
		h.List(c)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_GetByPublicID(t *testing.T) {
	t.Run("returns user when found", func(t *testing.T) {
		svc := &fakeService{
			user: &UserResponse{PublicID: "user1", Name: "Alice", Email: "alice@example.com", IsActive: true, RoleID: 1, RoleName: "admin", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		setTenant(c)
		h.GetByPublicID(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.APIResponse[UserResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Error("expected success response")
		}
		if resp.Data.PublicID != "user1" {
			t.Errorf("expected public_id 'user1', got %s", resp.Data.PublicID)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrNotFound}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		setTenant(c)
		h.GetByPublicID(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}

func TestHandler_Create(t *testing.T) {
	t.Run("creates user successfully", func(t *testing.T) {
		svc := &fakeService{
			created: &UserResponse{PublicID: "user3", Name: "Charlie", Email: "charlie@example.com", RoleID: 2, RoleName: "user"},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/users", CreateUserRequest{Name: "Charlie", Email: "charlie@example.com", Password: "Password1", RoleID: 2, CompanyID: uintPtr(7)})
		setTenant(c)
		h.Create(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var resp response.APIResponse[UserResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Error("expected success response")
		}
		if resp.Data.PublicID != "user3" {
			t.Errorf("expected public_id 'user3', got %s", resp.Data.PublicID)
		}
	})

	t.Run("passes tenant context to service", func(t *testing.T) {
		svc := &fakeService{created: &UserResponse{PublicID: "user3", Name: "Charlie", Email: "charlie@example.com", RoleID: 2, RoleName: "user"}}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/users", CreateUserRequest{Name: "Charlie", Email: "charlie@example.com", Password: "Password1", RoleID: 2})
		setTenant(c)
		h.Create(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
		}
		if svc.tenant != tenant.NewScoped(7, "acme") {
			t.Fatalf("expected tenant %#v, got %#v", tenant.NewScoped(7, "acme"), svc.tenant)
		}
	})

	t.Run("returns validation error on incomplete body", func(t *testing.T) {
		svc := &fakeService{}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/users", map[string]string{"email": "bad"})
		setTenant(c)
		h.Create(c)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", w.Code)
		}
	})

	t.Run("returns conflict on duplicate email", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrConflict}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/users", CreateUserRequest{Name: "Charlie", Email: "charlie@example.com", Password: "Password1", RoleID: 2, CompanyID: uintPtr(7)})
		setTenant(c)
		h.Create(c)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("returns validation error on invalid fields", func(t *testing.T) {
		svc := &fakeService{created: &UserResponse{PublicID: "user3", Name: "Charlie", Email: "charlie@example.com", RoleID: 2, RoleName: "user"}}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/users", CreateUserRequest{Name: "", Email: "not-email", Password: "short", RoleID: 0})
		setTenant(c)
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
		if _, ok := errsMap["email"]; !ok {
			t.Error("expected validation error for email")
		}
		if _, ok := errsMap["password"]; !ok {
			t.Error("expected validation error for password")
		}
		if _, ok := errsMap["role_id"]; !ok {
			t.Error("expected validation error for role_id")
		}
	})

	t.Run("passes authenticated actor public ID to service", func(t *testing.T) {
		svc := &fakeService{
			updated: &UserResponse{PublicID: "root1", Name: "Root", Email: "root@example.com", RoleID: 1, RoleName: "root"},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "root1"}}
		authctx.SetGin(c, &authctx.User{PublicID: "root1", IsActive: true})
		c.Request = jsonRequest(http.MethodPut, "/users/root1", UpdateUserRequest{Name: "Root", Email: "root@example.com", RoleID: RootRoleID})
		setTenant(c)
		h.Update(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if svc.updateActor != "root1" {
			t.Fatalf("expected actor 'root1', got %q", svc.updateActor)
		}
	})
}

func TestHandler_Update(t *testing.T) {
	t.Run("updates user successfully", func(t *testing.T) {
		svc := &fakeService{
			updated: &UserResponse{PublicID: "user1", Name: "Alice Updated", Email: "alice-new@example.com", RoleID: 2, RoleName: "user"},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPut, "/users/user1", UpdateUserRequest{Name: "Alice Updated", Email: "alice-new@example.com", RoleID: 2, CompanyID: uintPtr(7)})
		setTenant(c)
		h.Update(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.APIResponse[UserResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Data.Name != "Alice Updated" {
			t.Errorf("expected name 'Alice Updated', got %s", resp.Data.Name)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrNotFound}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		c.Request = jsonRequest(http.MethodPut, "/users/missing", UpdateUserRequest{Name: "Alice", Email: "alice@example.com", RoleID: 2, CompanyID: uintPtr(7)})
		setTenant(c)
		h.Update(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("returns conflict on duplicate email", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrConflict}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPut, "/users/user1", UpdateUserRequest{Name: "Alice", Email: "bob@example.com", RoleID: 2, CompanyID: uintPtr(7)})
		setTenant(c)
		h.Update(c)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("returns validation error on invalid fields", func(t *testing.T) {
		svc := &fakeService{updated: &UserResponse{PublicID: "user1", Name: "Alice", Email: "alice@example.com", RoleID: 2, RoleName: "user"}}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPut, "/users/user1", UpdateUserRequest{Name: "", Email: "bad", RoleID: 0})
		setTenant(c)
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
		if _, ok := errsMap["email"]; !ok {
			t.Error("expected validation error for email")
		}
		if _, ok := errsMap["role_id"]; !ok {
			t.Error("expected validation error for role_id")
		}
	})
}

func TestHandler_PassesTenantContextToService(t *testing.T) {
	svc := &fakeService{
		users: []UserResponse{{PublicID: "user1", Name: "Alice", Email: "alice@example.com", RoleID: 2}},
		total: 1,
	}
	_, h := setupHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users", nil)
	tenant.SetGin(c, tenant.NewScoped(42, "acme"))
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if svc.tenant.CompanyID != 42 || svc.tenant.IsRootScope {
		t.Fatalf("expected scoped tenant 42, got %#v", svc.tenant)
	}
}

func TestHandler_Delete(t *testing.T) {
	t.Run("deletes user successfully", func(t *testing.T) {
		svc := &fakeService{}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/users/user1", nil)
		setTenant(c)
		h.Delete(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if svc.deletedPID != "user1" {
			t.Errorf("expected deleted public_id 'user1', got %s", svc.deletedPID)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrNotFound}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/users/missing", nil)
		setTenant(c)
		h.Delete(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}

func TestHandler_ChangePassword(t *testing.T) {
	t.Run("changes password successfully", func(t *testing.T) {
		svc := &fakeService{}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPatch, "/users/user1/password", ChangePasswordRequest{CurrentPassword: "old", NewPassword: "NewPassword1"})
		setTenant(c)
		h.ChangePassword(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("returns unauthorized on wrong password", func(t *testing.T) {
		svc := &fakeService{changePasswordErr: apperror.ErrUnauthorized}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPatch, "/users/user1/password", ChangePasswordRequest{CurrentPassword: "wrong", NewPassword: "NewPassword1"})
		setTenant(c)
		h.ChangePassword(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		svc := &fakeService{changePasswordErr: apperror.ErrNotFound}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		c.Request = jsonRequest(http.MethodPatch, "/users/missing/password", ChangePasswordRequest{CurrentPassword: "old", NewPassword: "NewPassword1"})
		setTenant(c)
		h.ChangePassword(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("returns validation error on invalid fields", func(t *testing.T) {
		svc := &fakeService{}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPatch, "/users/user1/password", ChangePasswordRequest{CurrentPassword: "", NewPassword: "short"})
		setTenant(c)
		h.ChangePassword(c)

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
		if _, ok := errsMap["current_password"]; !ok {
			t.Error("expected validation error for current_password")
		}
		if _, ok := errsMap["new_password"]; !ok {
			t.Error("expected validation error for new_password")
		}
	})

	t.Run("passes authenticated actor public ID to password service", func(t *testing.T) {
		svc := &fakeService{}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "root1"}}
		authctx.SetGin(c, &authctx.User{PublicID: "root1", IsActive: true})
		c.Request = jsonRequest(http.MethodPatch, "/users/root1/password", ChangePasswordRequest{CurrentPassword: "old", NewPassword: "NewPassword1"})
		setTenant(c)
		h.ChangePassword(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if svc.passwordActor != "root1" {
			t.Fatalf("expected actor 'root1', got %q", svc.passwordActor)
		}
	})
}

func TestHandler_ToggleStatus(t *testing.T) {
	t.Run("toggles user status to inactive", func(t *testing.T) {
		svc := &fakeService{
			toggled: &UserResponse{PublicID: "user1", Name: "Alice", Email: "alice@example.com", IsActive: false, RoleID: 1, RoleName: "admin"},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPatch, "/users/user1/status", UpdateStatusRequest{IsActive: false})
		setTenant(c)
		h.ToggleStatus(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.APIResponse[UserResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Data.IsActive {
			t.Error("expected user to be inactive")
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrNotFound}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		c.Request = jsonRequest(http.MethodPatch, "/users/missing/status", UpdateStatusRequest{IsActive: false})
		setTenant(c)
		h.ToggleStatus(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}
