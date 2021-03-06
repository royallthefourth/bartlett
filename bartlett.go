// Package bartlett automatically generates an API from your database schema.
package bartlett

import (
	"database/sql"
	"net/http"
)

// A UserIDProvider is a function that is able to use an incoming request to produce a user ID.
type UserIDProvider func(r *http.Request) (interface{}, error)

// Bartlett holds all of the configuration necessary to generate an API from the database.
type Bartlett struct {
	DB     *sql.DB
	Driver Driver
	Tables []Table
	Users  UserIDProvider
}

func (b *Bartlett) ProbeTables(writable bool) *Bartlett {
	tables := b.Driver.ProbeTables(b.DB)
	for _, tbl := range tables {
		if !b.hasTable(tbl.Name) {
			tbl.Writable = writable
			b.Tables = append(b.Tables, tbl)
		}
	}

	return b
}

func (b *Bartlett) hasTable(name string) bool {
	for _, tbl := range b.Tables {
		if tbl.Name == name {
			return true
		}
	}
	return false
}
