package mdPoint

import (
	"context"
	"errors"
	"github.com/d3v-friends/go-rdbms/fnGorm"
	"github.com/d3v-friends/go-snippet/typ"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

var (
	ErrNotFoundAccount = errors.New("not found account")
)

type Account struct {
	Id        typ.UUID  `gorm:"primaryKey;type:char(36)"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time `gorm:"index"`

	// ref
	Coupons []*Coupon `gorm:"foreignKey:AccountId;references:Id"`
	Wallets []*Wallet `gorm:"foreignKey:AccountId;references:Id"`
}

func (x *Account) Migrate() []fnGorm.Migrate {
	return make([]fnGorm.Migrate, 0)
}

/*------------------------------------------------------------------------------------------------*/

type Accounts []*Account

func (x Accounts) FindById(id typ.UUID) (res *Account, err error) {
	for _, account := range x {
		if account.Id == id {
			res = account
			return
		}
	}
	err = ErrNotFoundAccount
	return
}

/*------------------------------------------------------------------------------------------------*/

// CreateAccount
// 새로운 계좌 생성
// 포인트 잔고 생성

type CreateAccountArgs struct {
	Fn func(ctx context.Context, i *Account) (err error)
}

func CreateAccount(ctx context.Context, i *CreateAccountArgs) error {
	return fnGorm.GetDBP(ctx).Transaction(func(tx *gorm.DB) (err error) {
		var now = time.Now()
		var account = &Account{
			Id:        typ.NewUUID(),
			CreatedAt: now,
			UpdatedAt: now,
		}
		var rows *gorm.DB
		if rows = tx.Create(account); rows.Error != nil {
			err = rows.Error
			return
		}

		var walletBalanceId = typ.NewUUID()
		var wallet = &Wallet{
			Id:              typ.NewUUID(),
			AccountId:       account.Id,
			WalletBalanceId: walletBalanceId,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		if rows = tx.
			Model(&Wallet{}).
			Create(wallet); rows.Error != nil {
			err = rows.Error
			return
		}

		account.Wallets = Wallets{wallet}

		var walletBalance = &WalletBalance{
			Id:           walletBalanceId,
			WalletId:     wallet.Id,
			Point:        decimal.Zero,
			PrevPoint:    decimal.Zero,
			ChangedPoint: decimal.Zero,
			Memo:         "",
			CreatedAt:    now,
		}

		if rows = tx.
			Model(&WalletBalance{}).
			Create(walletBalance); rows.Error != nil {
			err = rows.Error
			return
		}

		wallet.WalletBalance = walletBalance

		return i.Fn(ctx, account)
	})
}

/*------------------------------------------------------------------------------------------------*/

type FindAccountArgs struct {
	Id []typ.UUID
}

func (x FindAccountArgs) Query(db *gorm.DB) *gorm.DB {
	if len(x.Id) != 0 {
		db = db.Where("`accounts`.`id` IN (?)", x.Id)
	}

	db = db.
		Joins("Wallets").
		Preload("Wallets.WalletBalance").
		Joins("Coupons", func(tx *gorm.DB) *gorm.DB {
			return tx.
				Where("`coupons`.`has_balance` = ?", true).
				Order("`coupons`.`created_at` DESC")
		}).
		Preload("Coupons.CouponBalance")
	return db
}

func FindOneAccount(ctx context.Context, i *FindAccountArgs) (res *Account, err error) {
	var query = i.Query(fnGorm.GetDBP(ctx).Model(&Account{}))
	var rows *gorm.DB
	res = new(Account)

	if rows = query.Take(res); rows.Error != nil {
		err = rows.Error
		res = nil
		return
	}

	return
}

func FindAllAccount(ctx context.Context, i *FindAccountArgs) (res Accounts, err error) {
	res = make(Accounts, 0)
	var query = i.Query(fnGorm.GetDBP(ctx).Model(&Account{}))
	var rows *gorm.DB

	if rows = query.Find(&res); rows.Error != nil {
		err = rows.Error
		res = nil
		return
	}

	return
}
