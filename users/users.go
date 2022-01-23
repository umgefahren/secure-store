package users

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"
	"regexp"
	"secure-store/access"
)

const PasswordHashLength = 64
const PasswordSaltLength = 64

const APIKeyLength = 64

const ArgonTime = 1
const ArgonMemory = 64 * 1024
const ArgonThreads = 4

const nameExp = `^[a-zA-Z]+([ ]?[a-zA-Z]+)*$`
const usernameExp = `^[a-zA-Z]+([_]?[a-zA-Z]+)*$`

var nameMatcher = regexp.MustCompile(nameExp)
var usernameMatcher = regexp.MustCompile(usernameExp)

var NameIsAnIssue = errors.New("name doesn't match the requirement")
var UsernameIsAnIssue = errors.New("username doesn't match the requirement")
var PasswordHashHasWrongLength = errors.New("password hash has the wrong length")

type User struct {
	Id            uuid.UUID
	Name          string
	Role          Role
	Username      string
	PasswordHash  []byte
	PasswordSalt  []byte
	ApiKey        []byte
	BelongingKeys []*access.AKey
}

func (u *User) VerifyPassword(password []byte) bool {
	hash := argon2.IDKey(password, u.PasswordSalt, ArgonTime, ArgonMemory, ArgonThreads, PasswordHashLength)
	return bytes.Compare(hash, u.PasswordHash) == 0
}

func (u *User) VerifyApiKey(apiKey []byte) bool {
	return bytes.Compare(apiKey, u.ApiKey) == 0
}

type UserJson struct {
	Id           string `json:"Id"`
	Name         string `json:"Name"`
	Role         Role   `json:"Role"`
	Username     string `json:"Username"`
	PasswordHash []byte `json:"PasswordHash"`
	ApiKey       []byte `json:"ApiKey,omitempty"`
}

func (u *UserJson) IsValid() error {
	matchRes := nameMatcher.MatchString(u.Name)
	if !matchRes {
		return NameIsAnIssue
	}
	matchRes = usernameMatcher.MatchString(u.Username)
	if !matchRes {
		return UsernameIsAnIssue
	}
	if len(u.PasswordHash) != PasswordHashLength {
		return PasswordHashHasWrongLength
	}
	if len(u.ApiKey) != APIKeyLength {
		logrus.WithFields(logrus.Fields{
			"API Key":    hex.EncodeToString(u.ApiKey),
			"Key length": len(u.ApiKey),
		}).Infoln("api key length is wrong")
		return errors.New("api key length is wrong")
	}
	return nil
}

func UserFromUserJson(userJson *UserJson) (*User, error) {
	id, err := uuid.Parse(userJson.Id)
	if err != nil {
		return nil, err
	}
	salt := make([]byte, PasswordSaltLength)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, err
	}
	hashedPassword := argon2.IDKey(userJson.PasswordHash, salt, ArgonTime, ArgonMemory, ArgonThreads, PasswordHashLength)
	return &User{
		Id:            id,
		Name:          userJson.Name,
		Role:          userJson.Role,
		Username:      userJson.Username,
		PasswordHash:  hashedPassword,
		PasswordSalt:  salt,
		ApiKey:        userJson.ApiKey,
		BelongingKeys: nil,
	}, nil
}

type UserSafeJson struct {
	Id       string `json:"Id"`
	Name     string `json:"Name"`
	Role     Role   `json:"Role"`
	Username string `json:"Username"`
}

func UserSafeJsonFromUser(user *User) UserSafeJson {
	return UserSafeJson{
		Id:       user.Id.String(),
		Name:     user.Name,
		Role:     user.Role,
		Username: user.Username,
	}
}

type Role struct {
	RootUser       bool `json:"RootUser"`
	CanCreateUsers bool `json:"CanCreateUsers"`
	CanAddKeys     bool `json:"CanAddKeys"`
	CanUploadData  bool `json:"CanUploadData"`
	CanDeleteKeys  bool `json:"CanDeleteKeys"`
}

type UserStorage interface {
	Create(user *User) error
	ResolveByApiKey(apiKey []byte) (*User, error)
	ResolveByUsername(username string) (*User, error)
	ResolveByUuid(id uuid.UUID) (*User, error)
	Delete(id uuid.UUID) error
}

var UserAlreadyExists = errors.New("user already exists")
var UserDoesntExist = errors.New("user doesn't already exist")
var UserWithApiKeyAlreadyExists = errors.New("user with this api key already exists")
var UserWithApiKeyDoesntExist = errors.New("user with this api key doesn't exist")
var UserWithUsernameAlreadyExists = errors.New("user with this username already exists")
var UserWithUsernameDoesntExist = errors.New("user with this username doesn't exist")
