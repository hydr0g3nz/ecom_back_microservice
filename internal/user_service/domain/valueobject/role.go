package valueobject

import (
	"errors"
	"strings"
)

type Role string

const (
	Admin Role = "admin"
	User  Role = "user"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	roles := [...]Role{Admin, User}
	for _, role := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func ParseRole(role string) (Role, error) {
	role = strings.ToLower(role)
	if !Role(role).IsValid() {
		return "", errors.New("invalid role")
	}
	return Role(role), nil
}

func MustParseRole(role string) Role {
	r, err := ParseRole(role)
	if err != nil {
		panic(err)
	}
	return r
}
