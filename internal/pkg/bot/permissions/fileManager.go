package permissions

import (
	"encoding/json"
	"github.com/raf924/bot/internal/pkg/bot"
	"gopkg.in/yaml.v2"
	"os"
)

func init() {
	bot.Manage("yaml", newYamlFileManager)
	bot.Manage("json", newJsonFileManager)
}

type Decoder interface {
	Decode(v interface{}) error
}

type Encoder interface {
	Encode(v interface{}) error
}

type filePermissionManager struct {
	permissions map[string]bot.Permission
	encoder     Encoder
	decoder     Decoder
}

func (f *filePermissionManager) GetPermission(id string) (bot.Permission, error) {
	p, ok := f.permissions[id]
	if !ok {
		return bot.UNKNOWN, nil
	}
	return p, nil
}

func (f *filePermissionManager) SetPermission(id string, permission bot.Permission) error {
	f.permissions[id] = permission
	return f.encoder.Encode(f.permissions)
}

func newJsonFileManager(fileName string) bot.PermissionManager {
	f, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	encoder := json.NewEncoder(f)
	decoder := json.NewDecoder(f)
	var perms map[string]bot.Permission
	if err := decoder.Decode(&perms); err != nil {
		return nil
	}
	return &filePermissionManager{
		permissions: perms,
		encoder:     encoder,
		decoder:     decoder,
	}
}

func newYamlFileManager(fileName string) bot.PermissionManager {
	f, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	encoder := yaml.NewEncoder(f)
	decoder := yaml.NewDecoder(f)
	var perms map[string]bot.Permission
	if err := decoder.Decode(&perms); err != nil {
		return nil
	}
	return &filePermissionManager{
		permissions: perms,
		encoder:     encoder,
		decoder:     decoder,
	}
}
