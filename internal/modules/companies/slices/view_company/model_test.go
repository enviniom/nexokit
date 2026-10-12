package view_company

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"testing"
)

func TestCompany_TableName(t *testing.T) {
	c := company.Company{}
	if got := c.TableName(); got != "companies" {
		t.Errorf("company.Company.TableName() = %q, want %q", got, "companies")
	}
}

func TestCompanyDomain_TableName(t *testing.T) {
	d := company_domain.CompanyDomain{}
	if got := d.TableName(); got != "company_domains" {
		t.Errorf("company_domain.CompanyDomain.TableName() = %q, want %q", got, "company_domains")
	}
}
