package list_permissions

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

func TestPermissionResponseFromModel(t *testing.T) {
	actor := uint(42)
	now := time.Date(2025, 3, 4, 5, 6, 7, 0, time.UTC)
	model := permission.IAMPermission{
		BaseModel: shared.BaseModel{ID: 99, PublicID: "permission-public", CreatedAt: now, UpdatedAt: now.Add(time.Hour), CreatedBy: &actor, UpdatedBy: &actor, DeletedAt: gorm.DeletedAt{Time: now, Valid: true}},
		Slug:      "users.read", Name: "Read users", Module: "users", Action: "read", Description: "Read user records", IsSystem: true, DisplayOrder: 17,
	}
	want := permission.PermissionResponse{PublicID: model.PublicID, Slug: model.Slug, Name: model.Name, Module: model.Module, Action: model.Action, Description: model.Description, IsSystem: model.IsSystem, DisplayOrder: model.DisplayOrder, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt, CreatedBy: &actor, UpdatedBy: &actor}
	var response permission.PermissionResponse
	response.FromModel(model)
	if !reflect.DeepEqual(response, want) {
		t.Fatalf("mapping mismatch: got %#v, want %#v", response, want)
	}
	for _, tc := range []struct {
		name  string
		model permission.IAMPermission
		keys  int
	}{
		{"full public fields", model, 12}, {"zero model resets reused response and omits audit pointers", permission.IAMPermission{}, 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response.FromModel(tc.model)
			data, err := json.Marshal(response)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(data, &fields); err != nil {
				t.Fatal(err)
			}
			if len(fields) != tc.keys {
				t.Fatalf("unexpected fields: %s", data)
			}
			for _, key := range []string{"public_id", "slug", "name", "module", "action", "description", "is_system", "display_order", "created_at", "updated_at"} {
				if _, ok := fields[key]; !ok {
					t.Errorf("missing %s", key)
				}
			}
			for _, key := range []string{"id", "ID", "deleted_at", "DeletedAt"} {
				if _, ok := fields[key]; ok {
					t.Errorf("unsafe field %s", key)
				}
			}
			if tc.keys == 10 && !reflect.DeepEqual(response, permission.PermissionResponse{}) {
				t.Fatalf("stale response fields: %#v", response)
			}
		})
	}
}
