package permissions

import (
	botConfig "github.com/raf924/bot/pkg/config/bot"
	"github.com/raf924/connector-sdk/domain"
)

var permissionFormats = map[string]ManagerBuilder{}

type PermissionReader interface {
	GetPermission(id string) (domain.Permission, error)
}

type PermissionWriter interface {
	SetPermission(id string, permission domain.Permission) error
}

type PermissionManager interface {
	PermissionReader
	PermissionWriter
}

type ManagerBuilder func(location string) PermissionManager

func Manage(format string, builder ManagerBuilder) {
	println("Permission format:", format)
	permissionFormats[format] = builder
}

func GetManager(config botConfig.PermissionConfig) PermissionManager {
	builder, ok := permissionFormats[config.Format]
	if !ok {
		return nil
	}
	return builder(config.Location)
}

type noCheckPermissionManager struct {
}

func (n *noCheckPermissionManager) GetPermission(string) (domain.Permission, error) {
	return domain.IsAdmin, nil
}

func (n *noCheckPermissionManager) SetPermission(string, domain.Permission) error {
	return nil
}

func NewNoCheckPermissionManager() PermissionManager {
	return &noCheckPermissionManager{}
}
