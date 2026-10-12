package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/onboarding"
	"github.com/enviniom/nexokit/internal/modules/onboarding/core/company"
	"github.com/enviniom/nexokit/internal/modules/onboarding/core/company_domain"
	"github.com/enviniom/nexokit/internal/modules/onboarding/core/permission"
	"github.com/enviniom/nexokit/internal/modules/onboarding/core/role"
	"github.com/enviniom/nexokit/internal/modules/onboarding/core/user"
	"github.com/enviniom/nexokit/tests/helpers"
	"github.com/gin-gonic/gin"
)

type wiringHasher struct{}

func (wiringHasher) HashPassword(string) (string, error) { return "test-hash", nil }

// This verifies module wiring, not production authentication or PostgreSQL migrations.
func TestRegister_ProvisionsCompanyThroughContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	gin.SetMode(gin.TestMode)
	db := helpers.NewSQLiteDB(t, &company.OnboardingCompany{}, &company_domain.OnboardingCompanyDomain{}, &permission.OnboardingPermission{}, &role.OnboardingRole{}, &user.OnboardingUser{})
	if err := db.Exec("CREATE TABLE role_permissions (role_id integer not null, permission_id integer not null)").Error; err != nil {
		t.Fatal(err)
	}
	seed := permission.OnboardingPermission{Slug: "users.view", Name: "View users", Module: "users", Action: "view", IsSystem: true}
	seed.PublicID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	if err := db.Create(&seed).Error; err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	container := onboarding.NewContainer(db, onboarding.Config{PasswordHasher: wiringHasher{}, PlatformDomain: "platform.test"})
	guardCalls := 0
	onboarding.Register(router.Group("/api/v1"), container, func(required string) gin.HandlerFunc {
		if required != "root" {
			t.Fatalf("required role = %q", required)
		}
		return func(c *gin.Context) { guardCalls++; c.Next() }
	})
	body := `{"name":"Acme","slug":"acme","domain":"acme.test","generate_technical_domain":true,"admin_name":"Jane","admin_email":"jane@acme.test","admin_password":"SecurePass123"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/companies", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if guardCalls != 1 {
		t.Fatalf("guard calls=%d", guardCalls)
	}
	var envelope struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Data) != 4 || envelope.Data["company_slug"] != "acme" || envelope.Data["admin_email"] != "jane@acme.test" || len(envelope.Data["company_public_id"]) != 26 || len(envelope.Data["admin_public_id"]) != 26 {
		t.Fatalf("response=%v", envelope.Data)
	}
	for table, want := range map[string]int64{"companies": 1, "company_domains": 2, "roles": 2, "users": 1, "permissions": 1, "role_permissions": 2} {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("%s count=%d want=%d", table, count, want)
		}
	}
	var admin user.OnboardingUser
	if err := db.First(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if admin.PasswordHash != "test-hash" || admin.CompanyID == nil || admin.RoleID == 0 {
		t.Fatalf("administrator provisioning failed")
	}
}
