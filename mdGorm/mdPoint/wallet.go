package mdPoint

import (
	"context"
	"errors"
	"github.com/d3v-friends/go-rdbms/fnGorm"
	"github.com/d3v-friends/go-snippet/typ"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Wallet struct {
	Id              typ.UUID  `gorm:"primaryKey;type:char(36)"`
	AccountId       typ.UUID  `gorm:"index;type:char(36)"`
	WalletBalanceId typ.UUID  `gorm:"uniqueIndex;type:char(36)"`
	CreatedAt       time.Time `bson:"index"`
	UpdatedAt       time.Time `bson:"index"`

	// ref
	WalletBalance *WalletBalance `gorm:"foreignKey:WalletBalanceId;references:Id"`
}

func (x *Wallet) Migrate() []fnGorm.Migrate {
	return make([]fnGorm.Migrate, 0)
}

type WalletBalance struct {
	Id           typ.UUID        `gorm:"primaryKey;type:char(36)"`
	WalletId     typ.UUID        `gorm:"index;type:char(36)"`
	Point        decimal.Decimal `gorm:"type:decimal(20,2)"`
	PrevPoint    decimal.Decimal `gorm:"type:decimal(20,2)"`
	ChangedPoint decimal.Decimal `gorm:"type:decimal(20,2)"`
	Memo         string          `gorm:"type:text"`
	CreatedAt    time.Time       `gorm:"index"`

	// ref
	Wallet *Wallet `gorm:"foreignKey:Id;references:WalletId;"`
}

func (x *WalletBalance) Migrate() []fnGorm.Migrate {
	return make([]fnGorm.Migrate, 0)
}

type Wallets []*Wallet

/*------------------------------------------------------------------------------------------------*/

type FindWalletArgs struct {
	Id        []typ.UUID
	AccountId []typ.UUID
	Lock      *bool
}

func (x FindWalletArgs) Query(tx *gorm.DB) *gorm.DB {
	if len(x.Id) != 0 {
		tx = tx.Where("`wallets`.`id` IN (?)", x.Id)
	}

	if len(x.AccountId) != 0 {
		tx = tx.Where("`wallets`.`account_id` IN (?)", x.AccountId)
	}

	if x.Lock != nil && *x.Lock {
		tx = tx.Clauses(clause.Locking{
			Strength: "UPDATE",
		})
	}

	return tx
}

func FindOneWalletCtx(ctx context.Context, i *FindWalletArgs) (*Wallet, error) {
	return FindOneWallet(fnGorm.GetDBP(ctx), i)
}

func FindOneWallet(tx *gorm.DB, i *FindWalletArgs) (wallet *Wallet, err error) {
	wallet = new(Wallet)

	var rows *gorm.DB
	var query = tx.Model(&Wallet{})
	if rows = i.
		Query(query).
		Joins("WalletBalance").
		Take(wallet); rows.Error != nil {
		err = rows.Error
		return
	}

	return
}

func FindAllWallets(tx *gorm.DB, i *FindWalletArgs) (ls Wallets, err error) {
	var rows *gorm.DB
	ls = make(Wallets, 0)

	if rows = i.
		Query(tx.Model(&Wallet{})).
		Joins("WalletBalance").
		Order("`wallets`.`created_at` DESC").
		Find(&ls); rows.Error != nil {
		if errors.Is(rows.Error, gorm.ErrEmptySlice) {
			return
		}
		err = rows.Error
		return
	}

	return
}
