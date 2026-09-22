package constants

type Role string

const (
	RoleUser    Role = "user"
	RoleTeacher Role = "teacher"
	RoleAdmin   Role = "admin"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleTeacher, RoleAdmin:
		return true
	}
	return false
}

func (r Role) String() string {
	return string(r)
}
