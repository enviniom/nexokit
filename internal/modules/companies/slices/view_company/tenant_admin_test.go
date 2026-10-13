package view_company

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http/httptest"
	"testing"
)

func TestCompanyTenantIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&company.Company{}, &company_domain.CompanyDomain{}); err != nil {
		t.Fatal(err)
	}
	for _, item := range []company.Company{{BaseModel: shared.BaseModel{ID: 1, PublicID: "own"}, Name: "Own", Slug: "own"}, {BaseModel: shared.BaseModel{ID: 2, PublicID: "foreign"}, Name: "Foreign", Slug: "foreign"}} {
		if err = db.Create(&item).Error; err != nil {
			t.Fatal(err)
		}
	}
	id := uint(1)
	for _, tt := range []struct {
		name, target string
		root         bool
		status       int
	}{{"own company", "own", false, 200}, {"foreign company", "foreign", false, 404}, {"root global company", "foreign", true, 200}} {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.target}}
			c.Request = httptest.NewRequest("GET", "/companies/"+tt.target, nil)
			authctx.SetGin(c, &authctx.User{RoleSlug: "admin", CompanyID: &id, IsRoot: tt.root})
			if tt.root {
				tenant.SetGin(c, tenant.NewRoot())
			} else {
				tenant.SetGin(c, tenant.NewScoped(id, "own"))
			}
			NewHandler(NewService(NewRepository(db))).Handle(c)
			if w.Code != tt.status {
				t.Fatalf("want %d, got %d: %s", tt.status, w.Code, w.Body.String())
			}
		})
	}
}
