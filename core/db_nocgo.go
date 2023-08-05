//go:build !cgo

package core

import (
	"os"
	"strings"
	"text/template"

	_ "github.com/lib/pq"
	"github.com/pocketbase/dbx"
)

func connectDB(dbPath string) (*dbx.DB, error) {
	// Note: the busy_timeout pragma must be first because
	// the connection needs to be set to block on busy before WAL mode
	// is set in case it hasn't been already set by another connection.
	// pragmas := "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)"

	sentence := "host={{.hostt}} user={{.usert}} password={{.passwordt}} dbname={{.dbnamet}} port={{.portt}} sslmode=require"
	t, b := new(template.Template), new(strings.Builder)
	template.Must(t.Parse(sentence)).Execute(b, map[string]interface{}{
		"hostt":     os.Getenv("DB_HOST"),
		"usert":     os.Getenv("DB_USER"),
		"passwordt": os.Getenv("DB_PASSWORD"),
		"dbnamet":   os.Getenv("DB_NAME"),
		"portt":     os.Getenv("DB_PORT")})

	db, err := dbx.Open("postgres", b.String())
	if err != nil {
		return nil, err
	}

	return db, nil
}
