package companies

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/create_company_domain"
	"github.com/enviniom/nexokit/internal/modules/companies/delete_company"
	"github.com/enviniom/nexokit/internal/modules/companies/list_companies"
	"github.com/enviniom/nexokit/internal/modules/companies/list_company_domains"
	"github.com/enviniom/nexokit/internal/modules/companies/update_company"
	"github.com/enviniom/nexokit/internal/modules/companies/update_company_domain"
	"github.com/enviniom/nexokit/internal/modules/companies/view_company"
	"github.com/gin-gonic/gin"
)

func TestRegister_RouteAbsence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")
	container := &Container{
		ListCompanies:       &list_companies.Handler{},
		ViewCompany:         &view_company.Handler{},
		UpdateCompany:       &update_company.Handler{},
		DeleteCompany:       &delete_company.Handler{},
		ListCompanyDomains:  &list_company_domains.Handler{},
		CreateCompanyDomain: &create_company_domain.Handler{},
		UpdateCompanyDomain: &update_company_domain.Handler{},
	}
	Register(v1, container, func(string) gin.HandlerFunc { return func(c *gin.Context) { c.Next() } }, func(string) gin.HandlerFunc { return func(c *gin.Context) { c.Next() } })

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v1/companies"},
		{method: http.MethodDelete, path: "/api/v1/companies/01H/domains/01D"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != http.StatusNotFound {
				t.Fatalf("expected 404 for absent route, got %d", w.Code)
			}
		})
	}
}
