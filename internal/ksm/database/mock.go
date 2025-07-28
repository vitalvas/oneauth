package database

import (
	"fmt"

	"github.com/vitalvas/oneauth/internal/ksm/config"
)

// MockDB implements DB interface using in-memory SQLite for testing
// This provides more realistic testing than mocked functions while maintaining isolation
type MockDB struct {
	*SQLite
}

// NewMockDB creates a new MockDB instance with in-memory SQLite
func NewMockDB() (*MockDB, error) {
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	db, err := New(cfg)
	if err != nil {
		return nil, err
	}

	sqliteDB, ok := db.(*SQLite)
	if !ok {
		return nil, fmt.Errorf("expected SQLite database, got %T", db)
	}

	return &MockDB{SQLite: sqliteDB}, nil
}

// Reset clears all data from the MockDB (for test isolation)
func (m *MockDB) Reset() error {
	// Drop and recreate tables to clear all data
	_, err := m.db.Exec(`
		DROP TABLE IF EXISTS yubikey_counters;
		DROP TABLE IF EXISTS yubikey_keys;
	`)
	if err != nil {
		return err
	}

	// Recreate the schema
	return m.createTables()
}

// SetHealthCheckError allows tests to simulate database health check failures
// This is the only mocking behavior we preserve for specific test scenarios
type MockDBWithErrors struct {
	*MockDB
	healthCheckError error
}

// NewMockDBWithErrors creates a MockDB that can simulate specific error conditions
func NewMockDBWithErrors() (*MockDBWithErrors, error) {
	mockDB, err := NewMockDB()
	if err != nil {
		return nil, err
	}

	return &MockDBWithErrors{
		MockDB: mockDB,
	}, nil
}

// SetHealthCheckError configures the database to return an error on health checks
func (m *MockDBWithErrors) SetHealthCheckError(err error) {
	m.healthCheckError = err
}

// HealthCheck returns the configured error if set, otherwise performs real health check
func (m *MockDBWithErrors) HealthCheck() error {
	if m.healthCheckError != nil {
		return m.healthCheckError
	}
	return m.MockDB.HealthCheck()
}
