package main

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

func pointerString(s string) *string {
	return &s
}

func pointerInt(i int) *int {
	return &i
}

func pointerBool(b bool) *bool {
	return &b
}

func pointerTime(t time.Time) *time.Time {
	return &t
}

func pointerUUID(id uuid.UUID) *uuid.UUID {
	return &id
}
