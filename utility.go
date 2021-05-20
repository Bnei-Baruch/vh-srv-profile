package main

import uuid "github.com/satori/go.uuid"

func pointerString(s string) *string {
	return &s
}

func pointerUUID(id uuid.UUID) *uuid.UUID {
	return &id
}
