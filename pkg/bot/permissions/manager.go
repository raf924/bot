package permissions

import (
	botConfig "github.com/raf924/bot/pkg/config/bot"
)

var permissionFormats = map[string]ManagerBuilder{}

type PermissionReader interface {
	GetPermission(id string) (Permission, error)
}

type PermissionWriter interface {
	SetPermission(id string, permission Permission) error
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

func (n *noCheckPermissionManager) GetPermission(string) (Permission, error) {
	return ADMIN, nil
}

func (n *noCheckPermissionManager) SetPermission(string, Permission) error {
	return nil
}

func NewNoCheckPermissionManager() PermissionManager {
	return &noCheckPermissionManager{}
}
