package zaitun

import (
	"database/sql"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
)

type ZaitunController struct {
	db  *sql.DB
	res *lib.Responses
}

func NewZaitunController(db *sql.DB, res *lib.Responses) *ZaitunController {
	return &ZaitunController{db, res}
}
