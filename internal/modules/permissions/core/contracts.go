package core

// Resolver resolves permission slugs for an authenticated user public ID.
type Resolver interface {
	Resolve(publicID string) ([]string, error)
}

// Syncer synchronizes registered platform permissions into persistence.
type Syncer interface {
	SyncPermissions(slugs []string) error
}

// PermissionCatalogReader returns all permissions for roles catalog use-cases.
type PermissionCatalogReader interface {
	ListAll() ([]Permission, error)
}
