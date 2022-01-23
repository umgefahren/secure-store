package access

import "testing"

func TestAddAndDeleteMemory(t *testing.T) {
	db := NewMemoryStore()
	AddAndDeleteTest(t, db)
}

func TestAddAndAccessMemory(t *testing.T) {
	db := NewMemoryStore()
	AddAndAccessTest(t, db)
}
