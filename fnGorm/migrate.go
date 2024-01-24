package fnGorm

import (
	fnCases "github.com/d3v-friends/go-snippet/fn/fnCase"
	"github.com/d3v-friends/go-snippet/typ"
	"github.com/gertd/go-pluralize"
	"gorm.io/gorm"
	"reflect"
)

func RunMigrate(db *gorm.DB, models ...MigrateModel) error {
	return db.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.AutoMigrate(&Manage{}); err != nil {
			return
		}

		var ls = make(ManageList, 0)
		var res *gorm.DB
		if res = tx.Model(&Manage{}).Find(&ls); res.Error != nil {
			err = res.Error
			return
		}

		for _, md := range models {
			var tableNm = GetTableNm(md)
			if !ls.Has(tableNm) {
				if err = tx.AutoMigrate(md); err != nil {
					return
				}

				var manage = &Manage{
					Id:      typ.NewUUID(),
					TableNm: tableNm,
					NextIdx: 0,
				}

				if res = tx.Model(&Manage{}).Create(manage); res.Error != nil {
					err = res.Error
					return
				}

				ls = append(ls, manage)
			}

			var manage *Manage
			if manage, err = ls.Find(tableNm); err != nil {
				return
			}

			var fnMigrateList = md.Migrate()
			for i := manage.NextIdx; i < len(fnMigrateList); i++ {
				var fn = fnMigrateList[i]
				if err = fn(tx); err != nil {
					return
				}

				if res = tx.
					Model(&Manage{}).
					Where("`manage`.`table_nm` = ?", tableNm).
					Updates(map[string]interface{}{
						"next_idx": i + 1,
					}); res.Error != nil {
					return
				}
			}
		}

		return
	})
}

var plural = pluralize.NewClient()

func GetTableNm(v any) string {
	var t = reflect.TypeOf(v)
	return fnCases.SnakeCase(plural.Plural(t.Elem().Name()))
}
