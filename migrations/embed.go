// Package migrations embeds the SQL migration files so they ship inside the
// compiled binary and goose can run them without depending on the working directory.
package migrations

import "embed"

// FS holds all goose SQL migration files.
//
//go:embed *.sql
var FS embed.FS
