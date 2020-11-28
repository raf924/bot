package permissions

type Permission uint

const UNKNOWN Permission = 0
const (
	VERIFIED      Permission = 1
	NEED_VERIFIED Permission = 1
	MOD           Permission = 3
	NEED_MOD      Permission = 2
	ADMIN                    = 7
	NEED_ADMIN    Permission = 4
)

func (p Permission) Has(permission Permission) bool {
	return permission == 0 || p&permission != 0
}
