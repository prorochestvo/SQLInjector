package internal

import (
	"database/sql"
)

var _ Executor = &sql.DB{}
var _ Extractor = &sql.DB{}
var _ Vault = &sql.DB{}
