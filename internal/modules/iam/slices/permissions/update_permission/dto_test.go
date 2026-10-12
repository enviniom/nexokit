package update_permission

import (
	"reflect"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestUpdatePermissionRequestValidate(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request permission.UpdatePermissionRequest
		want    map[string][]string
	}{
		{name: "valid request", request: permission.UpdatePermissionRequest{Name: "Users Read"}, want: map[string][]string{}},
		{name: "missing name", want: map[string][]string{"name": {messages.MsgRequired}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.request.Validate(); !reflect.DeepEqual(map[string][]string(got), tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
