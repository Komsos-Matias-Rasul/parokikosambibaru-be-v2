package umkm

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func filterTokoByKategoriQuery(q string, kategori int) string {
	return fmt.Sprintf("%s WHERE kategori = %d", q, kategori)
}

func searchTokoQuery(q string, search string) string {
	return fmt.Sprintf("%s AND nama LIKE '%s'", q, search)
}

func sortTokoQuery(q string, sort string) string {
	if sort == "Z-A" {
		return fmt.Sprintf("%s ORDER BY nama DESC", q)
	}
	return fmt.Sprintf("%s ORDER BY nama ASC", q)
}

func (c *UMKMController) GetToko(ctx *gin.Context) {

	filter := ctx.Query("filter")
	search := ctx.Query("search")
	sort := ctx.Query("sort")

	type Toko struct {
		ID              *int    `json:"id"`
		Nama            *string `json:"nama"`
		Logo            *string `json:"logo"`
		Kategori        *int    `json:"kategori"`
		JenisProduk     *string `json:"jenisProduk"`
		Alamat          *string `json:"alamat"`
		KontakWhatsapp  *string `json:"whatsapp"`
		KontakInstagram *string `json:"instagram"`
		KontakFacebook  *string `json:"facebook"`
		KontakTelepon   *string `json:"telepon"`
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	q := `SELECT id, nama, logo, kategori, jenis_produk, alamat, kontak_whatsapp,
		kontak_instagram, kontak_facebook, kontak_telepon FROM umkm_toko`

	if filter != "" {
		kategori, _ := strconv.Atoi(filter)
		q = filterTokoByKategoriQuery(q, kategori)
	}
	if search != "" {
		q = searchTokoQuery(q, search)
	}
	q = sortTokoQuery(q, sort)
	tokoRows, err := c.db.QueryContext(_context, q)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	listToko := []*Toko{}
	for tokoRows.Next() {
		var toko Toko
		if err := tokoRows.Scan(
			&toko.ID,
			&toko.Nama,
			&toko.Logo,
			&toko.Kategori,
			&toko.JenisProduk,
			&toko.Alamat,
			&toko.KontakWhatsapp,
			&toko.KontakInstagram,
			&toko.KontakFacebook,
			&toko.KontakTelepon,
		); err != nil {
			log.Println(err.Error())
		}
		listToko = append(listToko, &toko)
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, gin.H{"toko": listToko})
}
