package umkm

import(
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (c *UMKMController) GetProductById(ctx *gin.Context){
	productId := ctx.Param("productId")

	type Toko struct {
		ID              *int    `json:"id"`
		Nama            *string `json:"nama"`
		Logo            *string `json:"logo"`
		Kategori        int     `json:"kategori"`
		KategoriLabel   string  `json:"kategoriLabel"`
		JenisProduk     *string `json:"jenisProduk"`
		Alamat          *string `json:"alamat"`
		KontakWhatsapp  *string `json:"whatsapp"`
		KontakInstagram *string `json:"instagram"`
		KontakFacebook  *string `json:"facebook"`
		KontakTelepon   *string `json:"telepon"`
	}

	type ProductDetail struct {
		ID			*int    `json:"id"`
		Name        *string `json:"name"`
		Img         *string `json:"img"`
		Price       *int    `json:"price"`
		Description *string `json:"description"`
		Toko        Toko    `json:"toko"`
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	q := `

		SELECT 
			up.id, up.nama, up.foto, up.harga, up.deskripsi,
			ut.id, ut.nama, ut.logo, ut.kategori, ut.jenis_produk,
			ut.alamat, ut.kontak_whatsapp, ut.kontak_instagram,
			ut.kontak_facebook, ut.kontak_telepon
		FROM umkm_products up
		JOIN umkm_toko ut ON up.id_toko = ut.id
		WHERE up.id = ?
	`

	var p ProductDetail
	var kategoriInt int

	err := c.db.QueryRowContext(_context, q, productId).Scan(
		&p.ID, &p.Name, &p.Img, &p.Price, &p.Description,
		&p.Toko.ID, &p.Toko.Nama, &p.Toko.Logo, &kategoriInt, &p.Toko.JenisProduk,
		&p.Toko.Alamat, &p.Toko.KontakWhatsapp, &p.Toko.KontakInstagram,
		&p.Toko.KontakFacebook, &p.Toko.KontakTelepon,
	)

	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err == sql.ErrNoRows {
		c.res.AbortWithStatusJSON(ctx, err, "produk tidak ditemukan", err.Error(), http.StatusNotFound, nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	p.Toko.Kategori = kategoriInt
	if kategoriInt >= 0 && kategoriInt < len(STORE_CATEGORY_MAPPING) {
		p.Toko.KategoriLabel = STORE_CATEGORY_MAPPING[kategoriInt]
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, gin.H{"product": p})
}