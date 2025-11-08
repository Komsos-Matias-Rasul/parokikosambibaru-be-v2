package profile

import (
	"database/sql"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
)

type ProfileController struct {
	db  *sql.DB
	res *lib.Responses
}

func NewProfileController(db *sql.DB, res *lib.Responses) *ProfileController {
	return &ProfileController{db, res}
}
