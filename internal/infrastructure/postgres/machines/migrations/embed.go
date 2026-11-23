// Package migrations embeds SQL migration files into the binary for use by
// cmd/migrate via the golang-migrate iofs driver.
package migrations

import "embed"

// FS holds all SQL migration files in this directory. It is used by
// cmd/migrate so the resulting binary is self-contained and does not rely on
// filesystem paths at runtime.
//
//go:embed *.sql
var FS embed.FS
