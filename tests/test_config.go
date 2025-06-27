package tests

import (
	"database/sql"
	"fmt"
	"internal-transfers/internal/config"
	"internal-transfers/tests/fixtures"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDatabaseConfig holds test database configuration
type TestDatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetTestDatabaseConfig returns test database configuration
func GetTestDatabaseConfig() *TestDatabaseConfig {
	return &TestDatabaseConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnv("TEST_DB_PORT", "5432"),
		User:     getEnv("TEST_DB_USER", "user"),
		Password: getEnv("TEST_DB_PASSWORD", "password"),
		DBName:   getEnv("TEST_DB_NAME", "internal_transfers_test"),
		SSLMode:  getEnv("TEST_DB_SSL_MODE", "disable"),
	}
}

// GetTestDatabaseURL returns the complete test database URL
func GetTestDatabaseURL() string {
	cfg := GetTestDatabaseConfig()
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
}

// SetupTestDatabase creates and configures a test database
func SetupTestDatabase(t *testing.T) *sql.DB {
	databaseURL := GetTestDatabaseURL()
	
	db, err := config.ConnectDB(databaseURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Create tables
	for _, query := range fixtures.DatabaseTestQueries.CreateTables {
		_, err := db.Exec(query)
		require.NoError(t, err, "Failed to create test table: %s", query)
	}

	return db
}

// CleanupTestDatabase removes all test data
func CleanupTestDatabase(t *testing.T, db *sql.DB) {
	for _, query := range fixtures.DatabaseTestQueries.CleanupData {
		_, err := db.Exec(query)
		if err != nil {
			t.Logf("Warning: Failed to cleanup with query %s: %v", query, err)
		}
	}
}

// SeedTestDatabase populates the database with test data
func SeedTestDatabase(t *testing.T, db *sql.DB) {
	for _, query := range fixtures.DatabaseTestQueries.SeedData {
		_, err := db.Exec(query)
		require.NoError(t, err, "Failed to seed test data: %s", query)
	}
}

// TeardownTestDatabase drops test tables and closes connection
func TeardownTestDatabase(t *testing.T, db *sql.DB) {
	// Clean up data first
	CleanupTestDatabase(t, db)
	
	// Drop tables
	for _, query := range fixtures.DatabaseTestQueries.DropTables {
		_, err := db.Exec(query)
		if err != nil {
			t.Logf("Warning: Failed to drop table with query %s: %v", query, err)
		}
	}
	
	db.Close()
}

// IsIntegrationTest checks if integration tests should run
func IsIntegrationTest() bool {
	return getEnv("RUN_INTEGRATION_TESTS", "false") == "true"
}

// RequireIntegrationTest skips test if integration tests are disabled
func RequireIntegrationTest(t *testing.T) {
	if !IsIntegrationTest() {
		t.Skip("Integration tests disabled. Set RUN_INTEGRATION_TESTS=true to enable.")
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// LogTestInfo logs test configuration information
func LogTestInfo() {
	log.Printf("Test Configuration:")
	log.Printf("  Database URL: %s", GetTestDatabaseURL())
	log.Printf("  Integration Tests: %t", IsIntegrationTest())
}

// TestEnvironment represents different test environments
type TestEnvironment string

const (
	EnvUnit        TestEnvironment = "unit"
	EnvIntegration TestEnvironment = "integration"
	EnvE2E         TestEnvironment = "e2e"
	EnvPerformance TestEnvironment = "performance"
)

// GetTestEnvironment returns the current test environment
func GetTestEnvironment() TestEnvironment {
	env := getEnv("TEST_ENV", "unit")
	return TestEnvironment(env)
}

// ShouldRunTest determines if a test should run in the current environment
func ShouldRunTest(t *testing.T, env TestEnvironment) bool {
	currentEnv := GetTestEnvironment()
	
	switch env {
	case EnvUnit:
		return true // Unit tests always run
	case EnvIntegration:
		return currentEnv == EnvIntegration || currentEnv == EnvE2E
	case EnvE2E:
		return currentEnv == EnvE2E
	case EnvPerformance:
		return currentEnv == EnvPerformance
	default:
		return true
	}
}

// SkipIfNotEnvironment skips test if not in specified environment
func SkipIfNotEnvironment(t *testing.T, env TestEnvironment) {
	if !ShouldRunTest(t, env) {
		t.Skipf("Test skipped in %s environment", GetTestEnvironment())
	}
}