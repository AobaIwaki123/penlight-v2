package seeds

import "embed"

// FS embeds seed SQL and master data json files.
//
//go:embed seed.sql data/*.json
var FS embed.FS
