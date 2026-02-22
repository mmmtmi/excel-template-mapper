package graph

import (
	"database/sql"

	"github.com/mmmtmi/excel-template-mapper/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	DB *sql.DB

	// Edit by CODEX: used by Excel upload mutations.
	Processor *service.Processor
}
