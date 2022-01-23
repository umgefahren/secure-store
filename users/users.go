package users

import (
	"github.com/google/uuid"
	"image"
	"secure-store/access"
)

type User struct {
	Id            uuid.UUID
	Name          string
	Role          Role
	Username      string
	PasswordHash  []byte
	ApiKeys       []byte
	BelongingKeys []*access.AKey
}

type Role struct {
	RootUser       bool
	CanCreateUsers bool
	CanAddKeys     bool
	CanUploadData  bool
}

type ProfileImage struct {
	Image     image.Image
	ImagePath string
}

type UserStorage interface {
	Create(user *User) error
	ResolveByApiKey(apiKey []byte) (*User, error)
	ResolveByUsername(username string) (*User, error)
	ResolveByUuid(id uuid.UUID) (*User, error)
	Delete(id uuid.UUID) error
}
