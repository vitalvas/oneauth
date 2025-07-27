package database

import (
	"github.com/stretchr/testify/mock"
)

// MockDB implements DB interface for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) StoreKey(key *YubikeyKey) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockDB) GetKey(keyID string) (*YubikeyKey, error) {
	args := m.Called(keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*YubikeyKey), args.Error(1)
}

func (m *MockDB) ListKeys() ([]*YubikeyKey, error) {
	args := m.Called()
	return args.Get(0).([]*YubikeyKey), args.Error(1)
}

func (m *MockDB) DeleteKey(keyID string) error {
	args := m.Called(keyID)
	return args.Error(0)
}

func (m *MockDB) UpdateKeyUsage(keyID string) error {
	args := m.Called(keyID)
	return args.Error(0)
}

func (m *MockDB) ValidateCounter(keyID string, counter, sessionUse int) error {
	args := m.Called(keyID, counter, sessionUse)
	return args.Error(0)
}

func (m *MockDB) StoreCounter(counter *YubikeyCounter) error {
	args := m.Called(counter)
	return args.Error(0)
}

func (m *MockDB) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}
