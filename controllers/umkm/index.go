package umkm

import (
	"database/sql"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
)

type UMKMController struct {
	db  *sql.DB
	res *lib.Responses
}

func NewUMKMController(db *sql.DB, res *lib.Responses) *UMKMController {
	return &UMKMController{db, res}
}
