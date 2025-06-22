package lib

import (
	"database/sql"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/conf"
	"github.com/go-sql-driver/mysql"
)

func GetDB() *sql.DB {
	conf := mysql.Config{
		User:         conf.MYSQL_USER,
		Passwd:       conf.MYSQL_PASSWORD,
		Addr:         conf.MYSQL_HOST,
		DBName:       conf.MYSQL_DATABASE,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
	}
	conn, err := mysql.NewConnector(&conf)
	if err != nil {
		panic(err.Error())
	}

	return sql.OpenDB(conn)
}
