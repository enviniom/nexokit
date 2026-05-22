package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/gin-gonic/gin"
)

func setupRecorder() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func parseBody(t *testing.T, w *httptest.ResponseRecorder) APIResponse[any] {
	t.Helper()
	var resp APIResponse[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return resp
}

func TestSuccess(t *testing.T) {
	c, w := setupRecorder()
	Success(c, "Operation completed", map[string]string{"status": "ok"})

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	resp := parseBody(t, w)
	if !resp.Success {
		t.Error("expected success to be true")
	}
	if resp.Message != "Operation completed" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
	if resp.Meta != nil {
		t.Error("expected meta to be nil")
	}
	if resp.Errors != nil {
		t.Error("expected errors to be nil")
	}
}

func TestCreated(t *testing.T) {
	c, w := setupRecorder()
	Created(c, "Resource created", map[string]int{"id": 1})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
	resp := parseBody(t, w)
	if !resp.Success {
		t.Error("expected success to be true")
	}
	if resp.Message != "Resource created" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
}

func TestNoContent(t *testing.T) {
	c, w := setupRecorder()
	NoContent(c)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", w.Body.String())
	}
}

func TestErrorResponse(t *testing.T) {
	c, w := setupRecorder()
	Error(c, http.StatusBadRequest, "Something went wrong", map[string][]string{"field": {"error"}})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	resp := parseBody(t, w)
	if resp.Success {
		t.Error("expected success to be false")
	}
	if resp.Message != "Something went wrong" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
	if resp.Data != nil {
		t.Error("expected data to be nil")
	}
}

func TestBadRequest(t *testing.T) {
	c, w := setupRecorder()
	BadRequest(c, "Invalid input")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	resp := parseBody(t, w)
	if resp.Message != "Invalid input" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
	if resp.Errors != nil {
		t.Error("expected errors to be nil")
	}
}

func TestNotFound(t *testing.T) {
	c, w := setupRecorder()
	NotFound(c, "Resource not found")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
	resp := parseBody(t, w)
	if resp.Message != "Resource not found" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
}

func TestUnauthorized(t *testing.T) {
	c, w := setupRecorder()
	Unauthorized(c, "Unauthorized access")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestForbidden(t *testing.T) {
	c, w := setupRecorder()
	Forbidden(c, "Access denied")
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestConflict(t *testing.T) {
	c, w := setupRecorder()
	Conflict(c, "Conflict detected")
	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestTooManyRequests(t *testing.T) {
	c, w := setupRecorder()
	TooManyRequests(c, messages.MsgTooManyRequests)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
	resp := parseBody(t, w)
	if resp.Message != messages.MsgTooManyRequests {
		t.Fatalf("message = %q; want %q", resp.Message, messages.MsgTooManyRequests)
	}
	if resp.Success {
		t.Fatal("expected success false")
	}
}

func TestInternalServerError(t *testing.T) {
	c, w := setupRecorder()
	InternalServerError(c, "Internal error")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestValidationError(t *testing.T) {
	c, w := setupRecorder()
	errs := map[string][]string{"email": {"must be a valid email"}}
	ValidationError(c, errs)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
	}
	resp := parseBody(t, w)
	if resp.Message != messages.MsgValidationError {
		t.Errorf("unexpected message: %s", resp.Message)
	}
	if resp.Errors == nil {
		t.Fatal("expected errors to be present")
	}
}

func TestPaginated(t *testing.T) {
	c, w := setupRecorder()
	data := []string{"a", "b", "c"}
	Paginated(c, "List retrieved", data, 1, 10, 25)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	resp := parseBody(t, w)
	if !resp.Success {
		t.Error("expected success to be true")
	}
	metaMap, ok := resp.Meta.(map[string]any)
	if !ok {
		t.Fatal("expected meta to be a map")
	}
	pgRaw, ok := metaMap["pagination"]
	if !ok {
		t.Fatal("expected pagination in meta")
	}
	pgJSON, _ := json.Marshal(pgRaw)
	var pg PaginationMeta
	if err := json.Unmarshal(pgJSON, &pg); err != nil {
		t.Fatalf("failed to unmarshal pagination: %v", err)
	}
	if pg.Page != 1 {
		t.Errorf("expected page 1, got %d", pg.Page)
	}
	if pg.PerPage != 10 {
		t.Errorf("expected per_page 10, got %d", pg.PerPage)
	}
	if pg.Total != 25 {
		t.Errorf("expected total 25, got %d", pg.Total)
	}
	if pg.TotalPages != 3 {
		t.Errorf("expected total_pages 3, got %d", pg.TotalPages)
	}
}

func TestPaginated_ZeroTotal(t *testing.T) {
	c, w := setupRecorder()
	Paginated(c, "Empty list", []string{}, 1, 10, 0)

	resp := parseBody(t, w)
	metaMap := resp.Meta.(map[string]any)
	pgRaw := metaMap["pagination"]
	pgJSON, _ := json.Marshal(pgRaw)
	var pg PaginationMeta
	_ = json.Unmarshal(pgJSON, &pg)
	if pg.TotalPages != 0 {
		t.Errorf("expected total_pages 0, got %d", pg.TotalPages)
	}
}

func TestRespondIfInvalid_NoErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errs := make(ValidationErrors)
	if RespondIfInvalid(c, errs) {
		t.Error("expected RespondIfInvalid to return false when no errors")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected no response written, got status %d", w.Code)
	}
}

func TestRespondIfInvalid_WithErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errs := make(ValidationErrors)
	errs.Add("email", messages.MsgValidEmail)
	if !RespondIfInvalid(c, errs) {
		t.Error("expected RespondIfInvalid to return true when errors exist")
	}
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
	}

	var resp APIResponse[any]
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Success {
		t.Error("expected success to be false")
	}
	if resp.Message != messages.MsgValidationError {
		t.Errorf("unexpected message: %s", resp.Message)
	}
}

func TestPaginatedWithFilters(t *testing.T) {
	tests := []struct {
		name       string
		params     query.ListParams
		wantStatus string
		wantSort   string
		wantOrder  string
		wantSearch string
	}{
		{
			name: "with filters and search",
			params: query.ListParams{
				Pagination: query.PaginationParams{Page: 1, PerPage: 10},
				Filters:    query.FilterParams{Status: "active"},
				Sort:       query.SortParams{Sort: "name", Order: "asc"},
				Search:     query.SearchParams{Query: "jhon"},
			},
			wantStatus: "active",
			wantSort:   "name",
			wantOrder:  "asc",
			wantSearch: "jhon",
		},
		{
			name:      "empty filters include defaults",
			params:    query.ListParams{Pagination: query.PaginationParams{Page: 2, PerPage: 25}},
			wantSort:  "created_at",
			wantOrder: "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupRecorder()
			PaginatedWithFilters(c, "List retrieved", []string{"a"}, tt.params, 50)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d; want %d", w.Code, http.StatusOK)
			}
			resp := parseBody(t, w)
			meta := resp.Meta.(map[string]any)
			filters := meta["filters"].(map[string]any)
			if filters["status"] != tt.wantStatus || filters["sort"] != tt.wantSort || filters["order"] != tt.wantOrder || filters["search"] != tt.wantSearch {
				t.Fatalf("filters = %#v; want status=%q sort=%q order=%q search=%q", filters, tt.wantStatus, tt.wantSort, tt.wantOrder, tt.wantSearch)
			}
			pagination := meta["pagination"].(map[string]any)
			if pagination["total"].(float64) != 50 {
				t.Fatalf("pagination.total = %#v; want 50", pagination["total"])
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{"not found", apperror.ErrNotFound, http.StatusNotFound, messages.MsgNotFound},
		{"validation", apperror.ErrValidation, http.StatusUnprocessableEntity, messages.MsgValidationError},
		{"unauthorized", apperror.ErrUnauthorized, http.StatusUnauthorized, messages.MsgUnauthorized},
		{"forbidden", apperror.ErrForbidden, http.StatusForbidden, messages.MsgForbidden},
		{"conflict", apperror.ErrConflict, http.StatusConflict, messages.MsgConflict},
		{"unknown", errors.New("database is down"), http.StatusInternalServerError, "database is down"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupRecorder()
			HandleError(c, tt.err)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d; want %d", w.Code, tt.wantStatus)
			}
			resp := parseBody(t, w)
			if resp.Message != tt.wantMessage {
				t.Fatalf("message = %q; want %q", resp.Message, tt.wantMessage)
			}
		})
	}
}

func TestHandleErrorNilIsNoOp(t *testing.T) {
	c, w := setupRecorder()
	HandleError(c, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want no response written", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("body = %q; want empty", w.Body.String())
	}
}
