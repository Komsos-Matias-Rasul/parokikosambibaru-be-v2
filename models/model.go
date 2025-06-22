package models

type Edition struct {
	Id          int     `json:"id"`
	Title       string  `json:"title"`
	CreatedAt   []uint8 `json:"createdAt"`
	PublishedAt []uint8 `json:"publishedAt"`
	EditionYear int     `json:"editionYear"`
	ArchivedAt  []uint8 `json:"archivedAt"`
	CoverImg    string  `json:"coverImg"`
	Thumbnail   string  `json:"thumbnail"`
}
