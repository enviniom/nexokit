package users

import (
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/users/assign_role_to_user"
	"github.com/enviniom/nexokit/internal/modules/iam/users/change_user_password"
	"github.com/enviniom/nexokit/internal/modules/iam/users/create_user"
	"github.com/enviniom/nexokit/internal/modules/iam/users/delete_user"
	"github.com/enviniom/nexokit/internal/modules/iam/users/list_users"
	"github.com/enviniom/nexokit/internal/modules/iam/users/toggle_user_status"
	"github.com/enviniom/nexokit/internal/modules/iam/users/update_user"
	"github.com/enviniom/nexokit/internal/modules/iam/users/view_user"
	"github.com/enviniom/nexokit/internal/platform/password"
	"gorm.io/gorm"
)

type Container struct {
	ListHandler           *list_users.Handler
	CreateHandler         *create_user.Handler
	ViewHandler           *view_user.Handler
	UpdateHandler         *update_user.Handler
	DeleteHandler         *delete_user.Handler
	ChangePasswordHandler *change_user_password.Handler
	AssignRoleHandler     *assign_role_to_user.Handler
	ToggleStatusHandler   *toggle_user_status.Handler
}

func NewContainer(db *gorm.DB, c cache.Cache) *Container {
	hasher := password.Manager{}

	return &Container{
		ListHandler:           list_users.NewHandler(list_users.NewService(list_users.NewRepository(db))),
		CreateHandler:         create_user.NewHandler(create_user.NewService(create_user.NewRepository(db), hasher)),
		ViewHandler:           view_user.NewHandler(view_user.NewService(view_user.NewRepository(db))),
		UpdateHandler:         update_user.NewHandler(update_user.NewService(update_user.NewRepository(db))),
		DeleteHandler:         delete_user.NewHandler(delete_user.NewService(delete_user.NewRepository(db))),
		ChangePasswordHandler: change_user_password.NewHandler(change_user_password.NewService(change_user_password.NewRepository(db), hasher)),
		AssignRoleHandler:     assign_role_to_user.NewHandler(assign_role_to_user.NewService(assign_role_to_user.NewRepository(db, c))),
		ToggleStatusHandler:   toggle_user_status.NewHandler(toggle_user_status.NewService(toggle_user_status.NewRepository(db))),
	}
}
