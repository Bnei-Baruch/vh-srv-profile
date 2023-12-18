package membership

import (
	"log"
)

func Migrate() {
	log.Println("Running membership migration")

	m := NewMigrator()

	if err := m.init(); err != nil {
		log.Fatalf("Error initializing migrator %s\n", err)
	}

	if err := m.migrate(); err != nil {
		log.Fatalf("Error migrating users to new membership model %s\n", err)
	}

	log.Println("Membership migration completed")
}

func Invalidate() {
	log.Println("Running membership invalidation")

	m := NewInvalidator()

	if err := m.init(); err != nil {
		log.Fatalf("Error initializing migrator %s\n", err)
	}

	if err := m.invalidate(); err != nil {
		log.Fatalf("Error migrating users to new membership model %s\n", err)
	}

	log.Println("Membership invalidation completed")
}
