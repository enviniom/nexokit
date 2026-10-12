package onboard_company

import (
	"encoding/json"
	"reflect"
	"testing"

	companyentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/company"
)

func TestWorkflowResponse_PublicSerialization(t *testing.T) {
	response := companyentity.OnboardCompanyResponse{CompanyPublicID: "company-public", CompanySlug: "acme", AdminPublicID: "admin-public", AdminEmail: "jane@acme.test"}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]string
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"company_public_id": "company-public", "company_slug": "acme", "admin_public_id": "admin-public", "admin_email": "jane@acme.test"}
	if !reflect.DeepEqual(fields, want) {
		t.Fatalf("public fields=%v want=%v", fields, want)
	}
}

func TestWorkflowRequest_ValidationDoesNotNormalize(t *testing.T) {
	domain := " ACME.TEST "
	request := companyentity.OnboardCompanyRequest{Name: " Acme ", Slug: " ACME ", Domain: &domain, AdminName: " Jane ", AdminEmail: "jane@acme.test", AdminPassword: "SecurePass123"}
	before := request
	if errs := request.Validate(); len(errs) > 0 {
		t.Fatalf("validation=%v", errs)
	}
	if !reflect.DeepEqual(request, before) || domain != " ACME.TEST " {
		t.Fatal("validation changed existing workflow normalization semantics")
	}
}
