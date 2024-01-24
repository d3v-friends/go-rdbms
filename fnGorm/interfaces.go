package fnGorm

import "gorm.io/gorm"

type (
	MigrateModel interface {
		Migrate() []Migrate
	}
	Migrate func(tx *gorm.DB) (err error)
)
