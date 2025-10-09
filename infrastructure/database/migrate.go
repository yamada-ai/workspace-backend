package database

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs database migrations
func RunMigrations(databaseURL, migrationsPath string) error {
	log.Println("🔄 Running database migrations...")

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Run migrations
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("✅ No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Migrations completed successfully")
	return nil
}
