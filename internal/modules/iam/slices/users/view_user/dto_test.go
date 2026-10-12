package view_user

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/shared"
)

func TestUserResponseFromModel(t *testing.T) {
	stamp := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	company, actor := uint(7), uint(9)
	for _, tt := range []struct {
		name           string
		company, actor *uint
		roleName       string
	}{
		{name: "optional fields omitted"},
		{name: "foreign and audit fields retained", company: &company, actor: &actor, roleName: "Admin"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			model := iamuser.IAMUser{BaseModel: shared.BaseModel{ID: 42, PublicID: "user-1", CreatedAt: stamp, UpdatedAt: stamp, CreatedBy: tt.actor, UpdatedBy: tt.actor}, Name: "Alice", Email: "alice@example.com", PasswordHash: "private-hash", RoleID: 20, CompanyID: tt.company, IsActive: true, Role: iamrole.IAMRole{Name: tt.roleName}}
			// A reused response must clear optional fields and the old role name.
			got := iamuser.UserResponse{CompanyID: &company, CreatedBy: &actor, UpdatedBy: &actor, RoleName: "stale"}
			got.FromModel(&model)
			want := iamuser.UserResponse{PublicID: model.PublicID, Name: model.Name, Email: model.Email, IsActive: model.IsActive, RoleID: model.RoleID, RoleName: tt.roleName, CompanyID: tt.company, CreatedAt: stamp, UpdatedAt: stamp, CreatedBy: tt.actor, UpdatedBy: tt.actor}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("mapping: got %#v, want %#v", got, want)
			}
			encoded, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"id", "password", "password_hash", "role"} {
				if _, ok := fields[key]; ok {
					t.Errorf("private field %s emitted", key)
				}
			}
			for _, key := range []string{"company_id", "created_by", "updated_by"} {
				if _, ok := fields[key]; ok != (tt.company != nil) {
					t.Errorf("optional field %s presence = %v", key, ok)
				}
			}
			if string(fields["role_id"]) != "20" {
				t.Fatalf("numeric role_id lost: %s", fields["role_id"])
			}
			if tt.company != nil && string(fields["company_id"]) != "7" {
				t.Fatalf("numeric company_id lost: %s", fields["company_id"])
			}
		})
	}
}
