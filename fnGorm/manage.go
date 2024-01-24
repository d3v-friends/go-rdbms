package fnGorm

import (
	"errors"
	"github.com/d3v-friends/go-snippet/typ"
	"time"
)

var (
	ErrNotFoundTable = errors.New("not found table")
)

type Manage struct {
	Id        typ.UUID  `gorm:"primaryKey;type:char(36)"`
	TableNm   string    `gorm:"uniqueIndex;type:varchar(50)"`
	NextIdx   int       `gorm:"type:int(8)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (x *Manage) Migrate() []Migrate {
	return []Migrate{}
}

type ManageList []*Manage

func (x *ManageList) Has(tableNm string) bool {
	for _, model := range *x {
		if model.TableNm == tableNm {
			return true
		}
	}
	return false
}

func (x *ManageList) Find(tableNm string) (res *Manage, err error) {
	for _, model := range *x {
		if model.TableNm == tableNm {
			res = model
			return
		}
	}
	err = ErrNotFoundTable
	return
}
