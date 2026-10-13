package view_company

import (
	"encoding/json"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"github.com/enviniom/nexokit/internal/shared"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

func companyRootContext(c *gin.Context) {
	authctx.SetGin(c, &authctx.User{IsRoot: true})
	tenant.SetGin(c, tenant.NewRoot())
}

type fakeViewService struct{ resp *company.CompanyResponse }

func (f *fakeViewService) GetByPublicID(tenant.TenantContext, string) (*company.CompanyResponse, error) {
	return f.resp, nil
}

func TestHandler_PublicProjection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	audit := uint(7)
	for _, withOptional := range []bool{true, false} {
		name := "without optional fields"
		if withOptional {
			name = "with audit and domains"
		}
		t.Run(name, func(t *testing.T) {
			model := &company.Company{BaseModel: shared.BaseModel{ID: 42, PublicID: "cmp_01"}, Name: "Acme", Slug: "acme", Status: "active"}
			if withOptional {
				model.CreatedBy, model.UpdatedBy = &audit, &audit
				model.Domains = []company_domain.CompanyDomain{{BaseModel: shared.BaseModel{ID: 99, PublicID: "dom_01"}, CompanyID: 42, Domain: "acme.com", Kind: "alias", Status: "active", RedirectToPrimary: true}}
			}
			r := gin.New()
			r.Use(companyRootContext)
			r.GET("/companies/:id", NewHandler(NewService(&fakeViewRepo{company: model})).Handle)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/companies/cmp_01", nil))
			var envelope struct {
				Data map[string]json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if w.Code != http.StatusOK || string(envelope.Data["public_id"]) != `"cmp_01"` {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if _, ok := envelope.Data["id"]; ok {
				t.Fatal("internal company ID exposed")
			}
			for _, key := range []string{"created_by", "updated_by", "domains"} {
				if _, ok := envelope.Data[key]; ok != withOptional {
					t.Fatalf("optional %s: %s", key, w.Body.String())
				}
			}
			if withOptional {
				var domains []map[string]json.RawMessage
				if err := json.Unmarshal(envelope.Data["domains"], &domains); err != nil {
					t.Fatal(err)
				}
				if len(domains) != 1 {
					t.Fatalf("domains=%v", domains)
				}
				for _, key := range []string{"id", "company_id"} {
					if _, ok := domains[0][key]; ok {
						t.Fatalf("internal domain field %s exposed", key)
					}
				}
				for key, want := range map[string]string{"public_id": `"dom_01"`, "company_public_id": `"cmp_01"`, "domain": `"acme.com"`, "redirect_to_primary": "true"} {
					if string(domains[0][key]) != want {
						t.Fatalf("domain %s=%s want %s", key, domains[0][key], want)
					}
				}
			}
		})
	}
}

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Authentication and access policy have already run at the route boundary.
	r.Use(func(c *gin.Context) { tenant.SetGin(c, tenant.NewScoped(7, "own")) })
	r.GET("/companies/:id", NewHandler(&fakeViewService{resp: &company.CompanyResponse{PublicID: "cmp_01"}}).Handle)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/companies/cmp_01", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
