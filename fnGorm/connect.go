package fnGorm

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"time"
)

type ConnectArgs struct {
	Host     string
	Username string
	Password string
	Schema   string
}

func (x *ConnectArgs) dsn() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true",
		x.Username,
		x.Password,
		x.Host,
		x.Schema,
	)
}

func NewConnect(i *ConnectArgs) (res *gorm.DB, err error) {
	if res, err = gorm.Open(
		mysql.New(mysql.Config{
			DSN: i.dsn(),
		}),
		&gorm.Config{
			FullSaveAssociations:                     false,
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger: logger.New(
				log.Default(),
				logger.Config{
					SlowThreshold:             time.Second, // Slow SQL threshold
					LogLevel:                  logger.Info, // Log level
					IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
					ParameterizedQueries:      true,        // Don't include params in the SQL log
					Colorful:                  true,        // Disable color
				}),
		},
	); err != nil {
		return
	}

	var sqlDb *sql.DB
	if sqlDb, err = res.DB(); err != nil {
		return
	}

	if err = sqlDb.Ping(); err != nil {
		return
	}

	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(50)
	sqlDb.SetConnMaxLifetime(5 * time.Minute)

	return
}

/*------------------------------------------------------------------------------------------------*/

type FnGorm struct {
	*gorm.DB
}

func (x *FnGorm) Migrate(models ...MigrateModel) error {
	return RunMigrate(x.DB, models...)
}
