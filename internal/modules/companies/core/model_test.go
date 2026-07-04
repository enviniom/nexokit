package core

import "testing"

func TestCompany_TableName(t *testing.T) {
	c := Company{}
	if got := c.TableName(); got != "companies" {
		t.Errorf("Company.TableName() = %q, want %q", got, "companies")
	}
}

func TestCompanyDomain_TableName(t *testing.T) {
	d := CompanyDomain{}
	if got := d.TableName(); got != "company_domains" {
		t.Errorf("CompanyDomain.TableName() = %q, want %q", got, "company_domains")
	}
}
