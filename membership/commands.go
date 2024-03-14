package membership

import (
	"log"
)

type doer interface {
	String() string
	init() error
	do() error
}

func Migrate() {
	do(NewMigrator())
}

func Invalidate() {
	do(NewInvalidator())
}

func BulkEval() {
	do(NewBulkEvaluator())
}

func do(doer doer) {
	log.Printf("Running doer: %s", doer)

	if err := doer.init(); err != nil {
		log.Fatalf("Error initializing doer %s: %s\n", doer, err)
	}

	if err := doer.do(); err != nil {
		log.Fatalf("Error doing %s: %s\n", doer, err)
	}

	log.Printf("%s completed\n", doer)
}
