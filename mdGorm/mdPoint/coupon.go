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

type Coupon struct {
	Id              typ.UUID        `gorm:"primaryKey;type:char(36)"`
	HasBalance      bool            `gorm:"index;"`                     // 잔고여부
	AccountId       typ.UUID        `gorm:"index;type:char(36);"`       // 소유권
	CouponBalanceId typ.UUID        `gorm:"uniqueIndex;type:char(36);"` // 잔고 정보
	Currency        decimal.Decimal `gorm:"type:decimal(20,2)"`         // 구매한 쿠폰 가격
	Price           decimal.Decimal `gorm:"type:decimal(20,2)"`         // 1건당 가격 예 15원
	Point           decimal.Decimal `gorm:"type:decimal(20,2)"`         // ( currency / price ) 갯수, 잔여 남으면 + 1건 해준다
	CreatedAt       time.Time       `gorm:"index;"`
	UpdatedAt       time.Time       `gorm:"index;"`

	// refer
	Account       *Account       `gorm:"foreignKey:AccountId;references:Id"`
	CouponBalance *CouponBalance `gorm:"foreignKey:CouponBalanceId;references:Id"`
}

func (x *Coupon) Migrate() []fnGorm.Migrate {
	return make([]fnGorm.Migrate, 0)
}

type CouponBalance struct {
	Id              typ.UUID        `gorm:"primaryKey;type:char(36)"`
	CouponId        typ.UUID        `gorm:"index"`
	Point           decimal.Decimal `gorm:"type:decimal(20,2)"`
	PrevPoint       decimal.Decimal `gorm:"type:decimal(20,2)"`
	ChangedPoint    decimal.Decimal `gorm:"type:decimal(20,2)"`
	Currency        decimal.Decimal `gorm:"type:decimal(20,2)"`
	PrevCurrency    decimal.Decimal `gorm:"type:decimal(20,2)"`
	ChangedCurrency decimal.Decimal `gorm:"type:decimal(20,2)"`
	CreatedAt       time.Time       `bson:"index"`
}

func (x *CouponBalance) Migrate() []fnGorm.Migrate {
	return make([]fnGorm.Migrate, 0)
}

/*------------------------------------------------------------------------------------------------*/

type Coupons []*Coupon

func (x Coupons) RestPoint() (point decimal.Decimal) {
	point = decimal.Zero
	for _, coupon := range x {
		point = point.Add(coupon.CouponBalance.Point)
	}
	return
}

func (x Coupons) RestCurrency() (currency decimal.Decimal) {
	currency = decimal.Zero
	for _, coupon := range x {
		currency = currency.Add(coupon.Price.Mul(coupon.CouponBalance.Point))
	}
	return
}

func (x Coupons) FilterByAccountId(id typ.UUID) (ls Coupons) {
	ls = make(Coupons, 0)
	for _, coupon := range x {
		if coupon.AccountId == id {
			ls = append(ls, coupon)
		}
	}
	return
}

/*------------------------------------------------------------------------------------------------*/

type CreateCouponArgs struct {
	AccountId typ.UUID
	Currency  decimal.Decimal
	Price     decimal.Decimal
	Fn        func(ctx context.Context, i *Coupon) (err error)
}

func (x CreateCouponArgs) Point() (point decimal.Decimal) {
	point = x.Currency.Div(x.Price)
	if point.Sub(point.Floor()).GreaterThan(decimal.Zero) {
		point = point.Floor().Add(decimal.NewFromInt(1))
	} else {
		point = point.Floor()
	}
	return
}

func CreateCoupon(ctx context.Context, i *CreateCouponArgs) error {
	return fnGorm.GetDBP(ctx).Transaction(func(tx *gorm.DB) (err error) {
		var couponBalanceId = typ.NewUUID()
		var now = time.Now()
		var coupon = &Coupon{
			Id:              typ.NewUUID(),
			HasBalance:      true,
			AccountId:       i.AccountId,
			CouponBalanceId: couponBalanceId,
			Currency:        i.Currency,
			Price:           i.Price,
			Point:           i.Point(),
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		var rows *gorm.DB
		if rows = tx.Model(&Coupon{}).Create(coupon); rows.Error != nil {
			err = rows.Error
			return
		}

		var adjustCurrency = coupon.Point.Mul(coupon.Price)
		var couponBalance = &CouponBalance{
			Id:              couponBalanceId,
			CouponId:        coupon.Id,
			Point:           coupon.Point,
			PrevPoint:       decimal.Zero,
			ChangedPoint:    coupon.Point,
			Currency:        adjustCurrency,
			PrevCurrency:    decimal.Zero,
			ChangedCurrency: adjustCurrency,
			CreatedAt:       now,
		}

		if rows = tx.Model(&CouponBalance{}).Create(couponBalance); rows.Error != nil {
			err = rows.Error
			return
		}

		coupon.CouponBalance = couponBalance

		return i.Fn(ctx, coupon)
	})
}

/*------------------------------------------------------------------------------------------------*/

type FindCouponArgs struct {
	AccountId  []typ.UUID
	HasBalance *bool
	Lock       *bool
}

func (x FindCouponArgs) Query(db *gorm.DB) *gorm.DB {
	if len(x.AccountId) != 0 {
		db = db.Where("`coupons`.`account_id` IN (?)", x.AccountId)
	}

	if x.HasBalance != nil {
		db = db.Where("`coupons`.`has_balance` = ?", *x.HasBalance)
	}

	if x.Lock != nil && *x.Lock {
		db = db.Clauses(clause.Locking{
			Strength: "UPDATE",
		})
	}

	return db
}

func FindAllCouponsWithCtx(ctx context.Context, i *FindCouponArgs) (Coupons, error) {
	var db = fnGorm.GetDBP(ctx)
	return FindAllCoupons(db, i)
}

func FindAllCoupons(tx *gorm.DB, i *FindCouponArgs) (res Coupons, err error) {
	var rows *gorm.DB
	res = make(Coupons, 0)

	var query = tx.Model(&Coupon{})
	if rows = i.Query(query).
		Joins("CouponBalance").
		Order("`coupons`.`created_at` DESC").
		Find(&res); rows.Error != nil {
		if errors.Is(rows.Error, gorm.ErrEmptySlice) {
			return
		}
		err = rows.Error
		res = nil
		return
	}
	return
}
