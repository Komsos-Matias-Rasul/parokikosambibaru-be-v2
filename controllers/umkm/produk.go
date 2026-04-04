package umkm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"database/sql"

	"github.com/gin-gonic/gin"
)

var CATEGORY_KULINER int = 1
var CATEGORY_FASHION int = 2
var CATEGORY_JASA int = 3
var CATEGORY_KERAJINAN int = 4
var CATEGORY_TEKNOLOGI_KREATIVITAS int = 5
var CATEGORY_KESEHATAN_PENDIDIKAN int = 6

var CATEGORY_KEYS []string = []string{
	"all",
	"kuliner",
	"fashion",
	"jasa",
	"kerajinan",
	"teknologi-kreativitas",
	"kesehatan-pendidikan",
}

var SORT_AZ int = 0
var SORT_ZA int = 1
var SORT_PRICE_ASC int = 2
var SORT_PRICE_DESC int = 3

var SORT_KEYS []string = []string{
	"az",
	"za",
	"price-asc",
	"price-desc",
}

func sortProdukQuery(q string, sort string) string {
	if sort == SORT_KEYS[SORT_ZA] {
		return fmt.Sprintf("%s ORDER BY up.nama DESC", q)
	}
	if sort == SORT_KEYS[SORT_PRICE_ASC] {
		return fmt.Sprintf("%s ORDER BY up.harga ASC", q)
	}
	if sort == SORT_KEYS[SORT_PRICE_DESC] {
		return fmt.Sprintf("%s ORDER BY up.harga DESC", q)
	}
	return fmt.Sprintf("%s ORDER BY up.nama ASC", q)
}

func searchProdukQuery(q string, search string) (string, string) {
	if search == "" {
		return q, ""
	}
	if strings.Contains(q, "WHERE") {
		return fmt.Sprintf("%s AND up.nama LIKE ?", q), "%" + search + "%"
	}
	return fmt.Sprintf("%s WHERE up.nama LIKE ?", q), "%" + search + "%"
}

func filterProductQuery(q string, category string) string {
	if category == CATEGORY_KEYS[CATEGORY_KULINER] {
		return fmt.Sprintf("%s WHERE ut.kategori=%d", q, CATEGORY_KULINER)
	}
	if category == CATEGORY_KEYS[CATEGORY_FASHION] {
		return fmt.Sprintf("%s WHERE ut.kategori=%d", q, CATEGORY_FASHION)
	}
	if category == CATEGORY_KEYS[CATEGORY_JASA] {
		return fmt.Sprintf("%s WHERE ut.kategori=%d", q, CATEGORY_JASA)
	}
	if category == CATEGORY_KEYS[CATEGORY_KERAJINAN] {
		return fmt.Sprintf("%s WHERE ut.kategori=%d", q, CATEGORY_KERAJINAN)
	}
	if category == CATEGORY_KEYS[CATEGORY_TEKNOLOGI_KREATIVITAS] {
		return fmt.Sprintf("%s WHERE ut.kategori=%d", q, CATEGORY_TEKNOLOGI_KREATIVITAS)
	}
	if category == CATEGORY_KEYS[CATEGORY_KESEHATAN_PENDIDIKAN] {
		return fmt.Sprintf("%s WHERE ut.kategori=%d", q, CATEGORY_KESEHATAN_PENDIDIKAN)
	}
	return q
}

func (c *UMKMController) GetProduct(ctx *gin.Context) {
	sort := ctx.Query("sort")
	if sort == "" {
		e := errors.New("invalid sort")
		c.res.AbortWithStatusJSON(ctx, e, "invalid sort", e.Error(), http.StatusBadRequest, nil)
		return
	}

	page := ctx.Query("page")
	category := ctx.Query("category")
	search := ctx.Query("q") 

	if category == "" {
		e := errors.New("invalid category")
		c.res.AbortWithStatusJSON(ctx, e, "invalid category", e.Error(), http.StatusBadRequest, nil)
		return
	}

	isValidCategory := slices.Index(CATEGORY_KEYS, category)
	if isValidCategory == -1 {
		e := errors.New("invalid category")
		c.res.AbortWithStatusJSON(ctx, e, "invalid category", e.Error(), http.StatusBadRequest, nil)
		return
	}

	currPage, err := strconv.Atoi(page)
	if err != nil {
		c.res.AbortWithStatusJSON(ctx, err, "invalid page", err.Error(), http.StatusBadRequest, nil)
		return
	}
	if currPage < 1 {
		e := errors.New("invalid page")
		c.res.AbortWithStatusJSON(ctx, e, "invalid page", e.Error(), http.StatusBadRequest, nil)
		return
	}

	type Product struct {
		ID        *int    `json:"id"`
		Name      *string `json:"name"`
		Img       *string `json:"img"`
		Price     *int    `json:"price"`
		StoreName *string `json:"storeName"`
		Category  string  `json:"category"`
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	
	qCount := `SELECT count(up.id) FROM umkm_products up 
        JOIN umkm_toko ut ON up.id_toko = ut.id`
	qCount = filterProductQuery(qCount, category)
	qCount, searchArg := searchProdukQuery(qCount, search)

	var ctr int
	if searchArg != "" {
		err = c.db.QueryRowContext(_context, qCount, searchArg).Scan(&ctr)
	} else {
		err = c.db.QueryRowContext(_context, qCount).Scan(&ctr)
	}
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	q := `SELECT up.id, up.nama, up.foto, up.harga, ut.nama, ut.kategori FROM umkm_products up 
        JOIN umkm_toko ut ON up.id_toko = ut.id`
	q = filterProductQuery(q, category)
	q, searchArg = searchProdukQuery(q, search)
	q = sortProdukQuery(q, sort)
	q = fmt.Sprintf("%s LIMIT 20 OFFSET %d", q, (currPage-1)*20)

	var produkRows *sql.Rows
	if searchArg != "" {
		produkRows, err = c.db.QueryContext(_context, q, searchArg)
	} else {
		produkRows, err = c.db.QueryContext(_context, q)
	}
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	listProduk := []*Product{}
	for produkRows.Next() {
		var produk Product
		var cat int
		if err := produkRows.Scan(
			&produk.ID, &produk.Name, &produk.Img,
			&produk.Price, &produk.StoreName, &cat,
		); err != nil {
			log.Println(err.Error())
		}
		produk.Category = STORE_CATEGORY_MAPPING[cat]
		listProduk = append(listProduk, &produk)
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, gin.H{
		"produk":    listProduk,
		"pageSize":  20,
		"currPage":  currPage,
		"itemCount": ctr,
		"pageCount": int(ctr/20) + 1,
	})
}

var MAXIMUM_OFFSET float32 = float32(433)
var STORE_CATEGORY_MAPPING []string = []string{
	"",
	"Kuliner",
	"Fashion",
	"Jasa",
	"Kerajinan",
	"Teknologi & Kreativitas",
	"Kesehatan & Pendidikan",
}

func (c *UMKMController) GetRandomProduct(ctx *gin.Context) {

	type Product struct {
		ID        *int    `json:"id"`
		Name      *string `json:"name"`
		Img       *string `json:"img"`
		Price     *int    `json:"price"`
		StoreName *string `json:"storeName"`
		Category  string  `json:"category"`
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	offset := int(rand.Float32() * MAXIMUM_OFFSET)
	produkRows, err := c.db.QueryContext(_context, `
		SELECT up.id, up.nama, up.foto, up.harga, ut.nama, ut.kategori FROM umkm_products up 
		JOIN umkm_toko ut ON up.id_toko = ut.id ORDER BY up.nama
		LIMIT 18 OFFSET ?
	`, offset)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	listProduk := []*Product{}
	for produkRows.Next() {
		var produk Product
		var c int
		if err := produkRows.Scan(
			&produk.ID,
			&produk.Name,
			&produk.Img,
			&produk.Price,
			&produk.StoreName,
			&c,
		); err != nil {
			log.Println(err.Error())
		}
		produk.Category = STORE_CATEGORY_MAPPING[c]
		listProduk = append(listProduk, &produk)
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, gin.H{"products": listProduk})
}
