package bot

type Permission string

const UNKNOWN Permission = "unknown"
const VERIFIED Permission = "verified"
const MOD Permission = "mod"
const ADMIN Permission = "admin"

type PermissionManager interface {
	GetPermission(id string) (Permission, error)
	SetPermission(id string, permission Permission) error
}

type ManagerBuilder func(location string) PermissionManager

var PermissionFormats = map[string]ManagerBuilder{}

func Manage(format string, builder ManagerBuilder) {
	PermissionFormats[format] = builder
}
