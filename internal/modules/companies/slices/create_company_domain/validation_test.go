package create_company_domain

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
)

func TestDomainValidationPersistsExistingNormalization(t *testing.T) {
	for _, input := range []string{"  ACME.COM  ", "acme.com"} {
		t.Run(input, func(t *testing.T) {
			create := company_domain.CreateCompanyDomainRequest{Domain: input, Kind: core.CompanyDomainKindAlias, Status: core.CompanyDomainStatusActive}
			if errs := create.Validate(); len(errs) != 0 {
				t.Fatalf("create validation: %v", errs)
			}
			if create.Domain != "acme.com" {
				t.Errorf("create domain = %q, want persisted normalization", create.Domain)
			}
			update := company_domain.UpdateCompanyDomainRequest{Domain: input, Kind: core.CompanyDomainKindAlias, Status: core.CompanyDomainStatusActive}
			if errs := update.Validate(); len(errs) != 0 {
				t.Fatalf("update validation: %v", errs)
			}
			if update.Domain != "acme.com" {
				t.Errorf("update domain = %q, want persisted normalization", update.Domain)
			}
		})
	}
}

func TestCompanyValidationPreservesNameAndSlug(t *testing.T) {
	create := company.CreateCompanyRequest{Name: "  Acme Inc  ", Slug: "  Acme-Slug  ", Status: core.CompanyStatusActive}
	if errs := create.Validate(); len(errs) != 0 {
		t.Fatalf("create validation: %v", errs)
	}
	if create.Name != "  Acme Inc  " || create.Slug != "  Acme-Slug  " {
		t.Fatalf("create unexpectedly normalized name/slug: %+v", create)
	}
	update := company.UpdateCompanyRequest{Name: create.Name, Slug: create.Slug, Status: core.CompanyStatusActive}
	if errs := update.Validate(); len(errs) != 0 {
		t.Fatalf("update validation: %v", errs)
	}
	if update.Name != create.Name || update.Slug != create.Slug {
		t.Fatalf("update unexpectedly normalized name/slug: %+v", update)
	}
}

func TestDomainValidationRejectsNormalizedEmptyAndPrimaryRedirect(t *testing.T) {
	for _, input := range []string{"   ", "acme.com"} {
		t.Run(input, func(t *testing.T) {
			create := company_domain.CreateCompanyDomainRequest{Domain: input, Kind: core.CompanyDomainKindPrimary, Status: core.CompanyDomainStatusActive, RedirectToPrimary: true}
			errs := create.Validate()
			if len(errs["redirect_to_primary"]) == 0 {
				t.Fatal("primary redirect must remain invalid")
			}
			if input == "   " && len(errs["domain"]) == 0 {
				t.Fatal("normalized empty domain must remain invalid")
			}
			update := company_domain.UpdateCompanyDomainRequest{Domain: input, Kind: create.Kind, Status: create.Status, RedirectToPrimary: true}
			errs = update.Validate()
			if len(errs["redirect_to_primary"]) == 0 {
				t.Fatal("update primary redirect must remain invalid")
			}
			if input == "   " && len(errs["domain"]) == 0 {
				t.Fatal("update normalized empty domain must remain invalid")
			}
		})
	}
}
