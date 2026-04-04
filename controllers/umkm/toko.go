package umkm

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func sortTokoQuery(q string, sort string) string {
	if sort == "Z-A" {
		return fmt.Sprintf("%s ORDER BY t.nama DESC", q)
	}
	return fmt.Sprintf("%s ORDER BY t.nama ASC", q)
}

func paginateQuery(q string, limit int, offset int) string {
	return fmt.Sprintf("%s LIMIT %d OFFSET %d", q, limit, offset)
}
	

func (c *UMKMController) GetToko(ctx *gin.Context) {

	filter := ctx.Query("filter")
	search := ctx.Query("q") 
	sort := ctx.Query("sort")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "15"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 15
	}
	offset := (page - 1) * limit

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
		JumlahProduk    int     `json:"jumlahProduk"`
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	baseQ := `SELECT t.id, t.nama, t.logo, t.kategori, t.jenis_produk, t.alamat,
		t.kontak_whatsapp, t.kontak_instagram, t.kontak_facebook, t.kontak_telepon,
		COUNT(p.id) AS jumlah_produk
		FROM umkm_toko t
		LEFT JOIN umkm_products p ON p.id_toko = t.id`

	countQ := `SELECT COUNT(DISTINCT t.id) FROM umkm_toko t`

	whereAdded := false
	if filter != "" {
		kategori, _ := strconv.Atoi(filter)
		baseQ = fmt.Sprintf("%s WHERE t.kategori = %d", baseQ, kategori)
		countQ = fmt.Sprintf("%s WHERE t.kategori = %d", countQ, kategori)
		whereAdded = true
	}
	if search != "" {
		if whereAdded {
			baseQ = fmt.Sprintf("%s AND t.nama LIKE '%%%s%%'", baseQ, search)
			countQ = fmt.Sprintf("%s AND t.nama LIKE '%%%s%%'", countQ, search)
		} else {
			baseQ = fmt.Sprintf("%s WHERE t.nama LIKE '%%%s%%'", baseQ, search)
			countQ = fmt.Sprintf("%s WHERE t.nama LIKE '%%%s%%'", countQ, search)
		}
	}

	baseQ = fmt.Sprintf("%s GROUP BY t.id", baseQ)

	baseQ = sortTokoQuery(baseQ, sort)
	baseQ = paginateQuery(baseQ, limit, offset)

	var total int
	if err := c.db.QueryRowContext(_context, countQ).Scan(&total); err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	tokoRows, err := c.db.QueryContext(_context, baseQ)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	defer tokoRows.Close()

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
			&toko.JumlahProduk,
		); err != nil {
			log.Println(err.Error())
		}
		listToko = append(listToko, &toko)
	}

	totalPages := (total + limit - 1) / limit

	c.res.SuccessWithStatusOKJSON(ctx, nil, gin.H{
		"toko":       listToko,
		"total":      total,
		"page":       page,
		"totalPages": totalPages,
	})
}





func (c *UMKMController) GetTokoById(ctx *gin.Context) {

	tokoId := ctx.Param("tokoId")

	type TokoDetail struct {
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

	type Produk struct {
		ID        *int    `json:"id"`
		Nama      *string `json:"name"`
		Harga     *int    `json:"price"`
		Deskripsi *string `json:"description"`
		Foto      *string `json:"img"`
	}






	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	var toko TokoDetail
	err := c.db.QueryRowContext(_context, `
		SELECT id, nama, logo, kategori, jenis_produk, alamat,
			kontak_whatsapp, kontak_instagram, kontak_facebook, kontak_telepon
		FROM umkm_toko
		WHERE id = ?
	`, tokoId).Scan(
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
	)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	produkRows, err := c.db.QueryContext(_context, `
		SELECT id, nama, harga, deskripsi, foto
		FROM umkm_products
		WHERE id_toko = ?
		ORDER BY nama ASC
	`, tokoId)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	defer produkRows.Close()

	listProduk := []*Produk{}
	for produkRows.Next() {
		var p Produk
		if err := produkRows.Scan(
			&p.ID,
			&p.Nama,
			&p.Harga,
			&p.Deskripsi,
			&p.Foto,
		); err != nil {
			log.Println(err.Error())
		}
		listProduk = append(listProduk, &p)
	}




	c.res.SuccessWithStatusOKJSON(ctx, nil, gin.H{
		"toko":   toko,
		"produk": listProduk,
	})
}