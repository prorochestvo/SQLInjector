package internal

import (
	"database/sql"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var _ Executor = &sql.DB{}
var _ Extractor = &sql.DB{}
var v Vault = &sql.DB{}

var _ boil.ContextExecutor = v
