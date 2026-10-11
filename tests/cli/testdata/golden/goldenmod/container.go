package goldenmod

import (
	create "github.com/enviniom/nexokit/internal/modules/goldenmod/slices/create_goldenmod"
	del "github.com/enviniom/nexokit/internal/modules/goldenmod/slices/delete_goldenmod"
	get "github.com/enviniom/nexokit/internal/modules/goldenmod/slices/get_goldenmod"
	list "github.com/enviniom/nexokit/internal/modules/goldenmod/slices/list_goldenmod"
	update "github.com/enviniom/nexokit/internal/modules/goldenmod/slices/update_goldenmod"
	"gorm.io/gorm"
)

// Container constructs dependencies only; app wiring and route guards remain explicit.
type Container struct {
	Create *create.Handler
	Get    *get.Handler
	List   *list.Handler
	Update *update.Handler
	Delete *del.Handler
}

// Enable GORM TranslateError in the application DB configuration to map
// database duplicate-key/foreign-key failures to the module's known errors.
func NewContainer(db *gorm.DB) *Container {
	return &Container{
		Create: create.NewHandler(create.NewRepository(db)),
		Get:    get.NewHandler(get.NewRepository(db)),
		List:   list.NewHandler(list.NewRepository(db)),
		Update: update.NewHandler(update.NewRepository(db)),
		Delete: del.NewHandler(del.NewRepository(db)),
	}
}
