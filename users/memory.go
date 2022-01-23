package users

import (
	"encoding/base64"
	"github.com/google/uuid"
	"sync"
)

type MemoryStore struct {
	userById       sync.Map
	uuidByApiKey   sync.Map
	uuidByUsername sync.Map
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		userById:       sync.Map{},
		uuidByApiKey:   sync.Map{},
		uuidByUsername: sync.Map{},
	}
}

func (m *MemoryStore) Create(user *User) error {
	id := user.Id
	_, ok := m.userById.Load(id)
	if ok {
		return UserAlreadyExists
	}

	_, ok = m.uuidByApiKey.Load(base64.StdEncoding.EncodeToString(user.ApiKey))
	if ok {
		return UserWithApiKeyAlreadyExists
	}
	_, ok = m.uuidByUsername.Load(user.Username)
	if ok {
		return UserWithUsernameAlreadyExists
	}
	m.userById.Store(id, user)
	m.uuidByApiKey.Store(base64.StdEncoding.EncodeToString(user.ApiKey), id)
	m.uuidByUsername.Store(user.Username, id)
	return nil
}

func (m *MemoryStore) ResolveByApiKey(apiKey []byte) (*User, error) {
	idInterface, ok := m.uuidByApiKey.Load(base64.StdEncoding.EncodeToString(apiKey))
	if !ok {
		return nil, UserWithApiKeyDoesntExist
	}
	id := idInterface.(uuid.UUID)
	return m.ResolveByUuid(id)
}

func (m *MemoryStore) ResolveByUsername(username string) (*User, error) {
	idInterface, ok := m.uuidByUsername.Load(username)
	if !ok {
		return nil, UserWithUsernameDoesntExist
	}
	id := idInterface.(uuid.UUID)
	return m.ResolveByUuid(id)
}

func (m *MemoryStore) ResolveByUuid(id uuid.UUID) (*User, error) {
	userInterface, ok := m.userById.Load(id)
	if !ok {
		return nil, UserDoesntExist
	}
	user := userInterface.(*User)
	return user, nil
}

func (m *MemoryStore) Delete(id uuid.UUID) error {
	_, ok := m.userById.Load(id)
	if !ok {
		return UserDoesntExist
	}
	m.userById.Delete(id)
	return nil
}
