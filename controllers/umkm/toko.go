package umkm

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func (c *UMKMController) GetToko(ctx *gin.Context) {

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
