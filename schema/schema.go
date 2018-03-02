package schema

import (
	"time"
)

type Item struct {
	Id       int       `db:"id" json:"id"`
	Price    int       `db:"price" json:"price"` // Only for New Taiwan Dollars, no cents
	Diff     int       `db:"diff" json:"diff"`
	Name     string    `db:"name" json:"name"`
	Category string    `db:"category" json:"category"`
	Url      string    `db:"url" json:"url"`
	Imgsrc   string    `db:"imgsrc" json:"imgsrc"`
	Source   string    `db:"source" json:"source"`
	Note     string    `db:"note" json:"note"`
	Created  time.Time `db:"created" json:"-"`
	Updated  time.Time `db:"updated" json:"updated,omitempty"`
}
