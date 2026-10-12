package view_role_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/shared"
)

func TestRoleResponseMapping(t *testing.T) {
	companyID, actorID := uint(12), uint(9)
	stamp := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	for _, tt := range []struct {
		name            string
		companyID       *uint
		companyPublicID string
		view            bool
		wantCompany     *string
	}{
		{name: "create update list omit absent company"},
		{name: "create update list omit empty preload", companyID: &companyID},
		{name: "view absent company", view: true},
		{name: "view preserves empty preload", companyID: &companyID, view: true, wantCompany: stringPointer("")},
		{name: "view ignores preload without foreign key", companyPublicID: "company", view: true},
		{name: "create update list populated company", companyID: &companyID, companyPublicID: "company", wantCompany: stringPointer("company")},
		{name: "view populated company", companyID: &companyID, companyPublicID: "company", view: true, wantCompany: stringPointer("company")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			model := iamrole.IAMRole{BaseModel: shared.BaseModel{PublicID: "role", CreatedAt: stamp, UpdatedAt: stamp, CreatedBy: &actorID, UpdatedBy: &actorID}, Name: "Reader", Slug: "reader", Description: "description", IsSystem: true, CompanyID: tt.companyID, Company: iamrole.IAMCompany{BaseModelSimple: shared.BaseModelSimple{PublicID: tt.companyPublicID}}}
			slugs := []string{"z.read", "a.read"}
			var got iamrole.RoleResponse
			if tt.view {
				got.FromViewModel(&model, slugs)
			} else {
				got.FromModel(&model, slugs)
			}
			want := iamrole.RoleResponse{PublicID: "role", Name: "Reader", Slug: "reader", Description: "description", IsSystem: true, CompanyID: tt.wantCompany, Permissions: slugs, CreatedAt: stamp, UpdatedAt: stamp, CreatedBy: &actorID, UpdatedBy: &actorID}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v want %#v", got, want)
			}
			data, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(data, &fields); err != nil {
				t.Fatal(err)
			}
			if _, present := fields["company_id"]; present != (tt.wantCompany != nil) {
				t.Fatalf("company omission: %s", data)
			}
		})
	}
}

func TestRoleMappingPreservesCallerOwnedCollections(t *testing.T) {
	for _, slugs := range [][]string{nil, {}, {"z", "a"}} {
		var got iamrole.RoleResponse
		got.FromModel(&iamrole.IAMRole{}, slugs)
		if !reflect.DeepEqual(got.Permissions, slugs) {
			t.Fatalf("permissions changed: %#v", got.Permissions)
		}
		var assignment iamrole.RolePermissionAssignmentResponse
		assignment.FromSelection("role", slugs, nil)
		if assignment.RoleID != "role" || !reflect.DeepEqual(assignment.Permissions, slugs) || assignment.Catalog != nil {
			t.Fatalf("assignment changed: %#v", assignment)
		}
	}
	var got iamrole.RoleResponse
	got.FromModel(&iamrole.IAMRole{}, nil)
	if got.CreatedBy != nil || got.UpdatedBy != nil {
		t.Fatalf("optional audit fields populated: %#v", got)
	}
}

func TestRolePermissionCatalogMapping(t *testing.T) {
	items := []permission.IAMPermission{
		{BaseModel: shared.BaseModel{PublicID: "first"}, Slug: "z.read", Name: "Read", Module: "z", Action: "read", Description: "description", IsSystem: true, DisplayOrder: 7},
		{BaseModel: shared.BaseModel{PublicID: "second"}, Slug: "a.write", Module: "a"},
		{BaseModel: shared.BaseModel{PublicID: "third"}, Slug: "z.write", Module: "z"},
	}
	got := iamrole.BuildPermissionCatalog(items, map[string]bool{"z.read": true, "unknown": true})
	want := []iamrole.RolePermissionGroupResponse{
		{Module: "z", Permissions: []iamrole.RolePermissionResponse{
			{PublicID: "first", Slug: "z.read", Name: "Read", Module: "z", Action: "read", Description: "description", IsSystem: true, DisplayOrder: 7, Granted: true},
			{PublicID: "third", Slug: "z.write", Module: "z"},
		}},
		{Module: "a", Permissions: []iamrole.RolePermissionResponse{{PublicID: "second", Slug: "a.write", Module: "a"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	for _, items := range [][]permission.IAMPermission{nil, {}} {
		got := iamrole.BuildPermissionCatalog(items, nil)
		if got == nil || len(got) != 0 {
			t.Fatalf("empty catalog must be []: %#v", got)
		}
	}
	var projection iamrole.RolePermissionResponse
	projection.FromModel(&items[0], nil)
	if projection.Granted {
		t.Fatal("nil grants must deny")
	}
}

func stringPointer(value string) *string { return &value }
