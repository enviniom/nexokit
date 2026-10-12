package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies"
	"github.com/enviniom/nexokit/internal/modules/companies/slices/create_company_domain"
	"github.com/enviniom/nexokit/internal/modules/companies/slices/delete_company"
	"github.com/enviniom/nexokit/internal/modules/companies/slices/list_companies"
	"github.com/enviniom/nexokit/internal/modules/companies/slices/list_company_domains"
	"github.com/enviniom/nexokit/internal/modules/companies/slices/update_company"
	"github.com/enviniom/nexokit/internal/modules/companies/slices/update_company_domain"
	"github.com/enviniom/nexokit/internal/modules/companies/slices/view_company"
	"github.com/gin-gonic/gin"
)

// Company creation belongs to onboarding, not the companies module.
func TestCompaniesRegister_CreationBelongsToOnboarding(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	container := &companies.Container{
		ListCompanies:       &list_companies.Handler{},
		ViewCompany:         &view_company.Handler{},
		UpdateCompany:       &update_company.Handler{},
		DeleteCompany:       &delete_company.Handler{},
		ListCompanyDomains:  &list_company_domains.Handler{},
		CreateCompanyDomain: &create_company_domain.Handler{},
		UpdateCompanyDomain: &update_company_domain.Handler{},
	}
	guard := func(string) gin.HandlerFunc { return func(c *gin.Context) { c.Next() } }
	companies.Register(r.Group("/api/v1"), container, guard, guard)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/companies", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("company creation must use onboarding: status=%d body=%s", w.Code, w.Body.String())
	}
}
