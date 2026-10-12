package view_company_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"github.com/enviniom/nexokit/internal/shared"
	"github.com/enviniom/nexokit/tests/helpers"
)

func TestCompanyResponseMappingAndSerialization(t *testing.T) {
	audit := uint(7)
	stamp := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	domain := company_domain.CompanyDomain{BaseModel: shared.BaseModel{ID: 99, PublicID: "domain-public", CreatedAt: stamp, UpdatedAt: stamp}, Domain: "acme.test", Status: "active", Kind: "alias", RedirectToPrimary: true}
	model := company.Company{BaseModel: shared.BaseModel{ID: 42, PublicID: "company-public", CreatedAt: stamp, UpdatedAt: stamp, CreatedBy: &audit, UpdatedBy: &audit}, Name: "Acme", Slug: "acme", Status: "active", Domains: []company_domain.CompanyDomain{domain}}
	for _, tt := range []struct {
		name    string
		include bool
		domains []company_domain.CompanyDomain
	}{
		{"view includes domains", true, model.Domains},
		{"list update exclude domains", false, model.Domains},
		{"nil domains omitted", true, nil},
		{"empty domains omitted", true, []company_domain.CompanyDomain{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := model
			m.Domains = tt.domains
			var dto company.CompanyResponse
			dto.FromModel(m, tt.include)
			if dto.PublicID != m.PublicID || dto.Name != m.Name || dto.Slug != m.Slug || dto.Status != m.Status || !dto.CreatedAt.Equal(stamp) || !dto.UpdatedAt.Equal(stamp) || dto.CreatedBy != &audit || dto.UpdatedBy != &audit {
				t.Fatalf("lost company fields: %+v", dto)
			}
			wantDomains := tt.include && len(tt.domains) > 0
			if (len(dto.Domains) > 0) != wantDomains {
				t.Fatalf("domain projection: %+v", dto)
			}
			if wantDomains {
				var expected company_domain.CompanyDomainResponse
				expected.FromModel(domain, m.PublicID)
				if !reflect.DeepEqual(dto.Domains[0], expected) || expected.CompanyPublicID != m.PublicID || expected.PublicID != domain.PublicID || expected.Domain != domain.Domain || expected.Kind != domain.Kind || expected.Status != domain.Status || !expected.RedirectToPrimary || !expected.CreatedAt.Equal(stamp) || !expected.UpdatedAt.Equal(stamp) {
					t.Fatalf("lost domain fields: %+v", expected)
				}
			}
			data, err := json.Marshal(dto)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err = json.Unmarshal(data, &fields); err != nil {
				t.Fatal(err)
			}
			if _, ok := fields["id"]; ok {
				t.Fatal("private ID exposed")
			}
			if _, ok := fields["domains"]; ok != wantDomains {
				t.Fatalf("domains JSON shape: %s", data)
			}
			for _, key := range []string{"public_id", "name", "slug", "status", "created_at", "updated_at", "created_by", "updated_by"} {
				if _, ok := fields[key]; !ok {
					t.Errorf("missing %s", key)
				}
			}
			dto.FromModel(company.Company{}, false)
			data, err = json.Marshal(dto)
			if err != nil {
				t.Fatal(err)
			}
			fields = nil
			if err = json.Unmarshal(data, &fields); err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"domains", "created_by", "updated_by"} {
				if _, ok := fields[key]; ok {
					t.Errorf("stale optional %s", key)
				}
			}
		})
	}
}

func TestCompanyDomainAssociationLoading(t *testing.T) {
	if got := (company_domain.CompanyAssociation{}).TableName(); got != "companies" {
		t.Fatalf("association table %q", got)
	}
	db := helpers.NewSQLiteDB(t, &company.Company{}, &company_domain.CompanyDomain{})
	c := company.Company{BaseModel: shared.BaseModel{PublicID: "company-public"}, Name: "Acme", Slug: "acme", Status: "active"}
	if err := db.Create(&c).Error; err != nil {
		t.Fatal(err)
	}
	d := company_domain.CompanyDomain{BaseModel: shared.BaseModel{PublicID: "domain-public"}, CompanyID: c.ID, Domain: "acme.test", Status: "active", Kind: "primary"}
	if err := db.Create(&d).Error; err != nil {
		t.Fatal(err)
	}
	var loaded company_domain.CompanyDomain
	if err := db.Preload("Company").First(&loaded, d.ID).Error; err != nil {
		t.Fatal(err)
	}
	if loaded.Company.ID != c.ID || loaded.Company.PublicID != c.PublicID || loaded.Company.Name != c.Name || loaded.Company.Slug != c.Slug || loaded.Company.Status != c.Status {
		t.Fatalf("association not loaded: %+v", loaded.Company)
	}
	var parent company.Company
	if err := db.Preload("Domains").First(&parent, c.ID).Error; err != nil {
		t.Fatal(err)
	}
	if len(parent.Domains) != 1 || parent.Domains[0].PublicID != d.PublicID {
		t.Fatalf("domains not loaded: %+v", parent.Domains)
	}
}
