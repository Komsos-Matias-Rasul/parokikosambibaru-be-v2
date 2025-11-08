package editor

import (
	"database/sql"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
)

type EditorController struct {
	db  *sql.DB
	res *lib.Responses
}

func NewEditorController(db *sql.DB, res *lib.Responses) *EditorController {
	return &EditorController{db, res}
}
