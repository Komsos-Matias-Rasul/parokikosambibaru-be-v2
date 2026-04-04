package umkm


import(
	"context"
	"time"

	"log"

	"github.com/gin-gonic/gin"
)

func (c *UMKMController) GetSuggest (ctx *gin.Context){
	q := ctx.Query("q")

	if len(q) < 2 {
		c.res.SuccessWithStatusOKJSON(ctx, nil, gin.H{
			"produk": []any{},
			"toko": []any{},
			
		})
		return
	}

	type ProdukSuggest struct {
		ID		*int	`json:"id"`
		Nama	*string `json:"nama"`
	}

	type TokoSuggest struct {
		ID		*int `json:"id"`
		Nama	*string `json:"nama"`
		Logo	*string `json:"logo"`
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	like := "%" + q + "%"

	produkRows, err := c.db.QueryContext(_context, `
		SELECT 
			id, nama
		from 
			umkm_products
		Where
			nama 
		LIKE 
			?
		order by
			nama ASC
		limit 5


	`, like)

	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	defer produkRows.Close()
 
	listProduk := []ProdukSuggest{}
	for produkRows.Next() {
		var p ProdukSuggest
		if err := produkRows.Scan(&p.ID, &p.Nama); err != nil {
			log.Println(err.Error())
			continue
		}
		listProduk = append(listProduk, p)
	}

	tokoRows, err := c.db.QueryContext(_context, `
		SELECT id, nama, logo
		FROM umkm_toko
		WHERE nama LIKE ?
		ORDER BY nama ASC
		LIMIT 3
	`, like)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	defer tokoRows.Close()
 
	listToko := []TokoSuggest{}
	for tokoRows.Next() {
		var t TokoSuggest
		if err := tokoRows.Scan(&t.ID, &t.Nama, &t.Logo); err != nil {
			log.Println(err.Error())
			continue
		}
		listToko = append(listToko, t)
	}
 
	c.res.SuccessWithStatusOKJSON(ctx, nil, gin.H{
		"produk": listProduk,
		"toko":   listToko,
	})

};