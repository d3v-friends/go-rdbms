package mdPoint

import (
	"context"
	"errors"
	"fmt"
	"github.com/d3v-friends/go-rdbms/fnGorm"
	"github.com/d3v-friends/go-snippet/fn/fnReflect"
	"github.com/d3v-friends/go-snippet/typ"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

var (
	ErrShortagePoint = errors.New("shortage point")
)

type CouponUseRequest struct {
	Id           typ.UUID        `gorm:"primaryKey;type:char(36)"`
	AccountId    typ.UUID        `gorm:"index;type:char(36)"`
	UsedPoint    decimal.Decimal `gorm:"decimal(20,2)"`
	UsedCurrency decimal.Decimal `gorm:"decimal(20,2)"`
	Msg          string          `gorm:"type:text"`
	CreatedAt    time.Time       `gorm:"index"`

	// ref
	CouponUseReceipts []*CouponUseReceipt `gorm:"foreignKey:CouponUseRequestId;references:Id"`
}

func (x *CouponUseRequest) Migrate() []fnGorm.Migrate {
	return make([]fnGorm.Migrate, 0)
}

type CouponUseReceipt struct {
	Id                 typ.UUID `gorm:"primaryKey;type:char(36)"`
	CouponUseRequestId typ.UUID `gorm:"index;type:char(36)"`
	CouponBalanceId    typ.UUID `gorm:"index;type:char(36)"`

	// ref
	CouponUseRequest *CouponUseRequest `gorm:"foreignKey:CouponUseRequestId;references:Id"`
	CouponBalance    *CouponBalance    `gorm:"foreignKey:CouponBalanceId;references:Id"`
}

func (x *CouponUseReceipt) Migrate() []fnGorm.Migrate {
	return make([]fnGorm.Migrate, 0)
}

/*------------------------------------------------------------------------------------------------*/

type UseCouponArgs struct {
	AccountId typ.UUID
	Point     decimal.Decimal
	Fn        func(ctx context.Context, req *CouponUseRequest) (err error)
}

func UseCoupon(ctx context.Context, i *UseCouponArgs) error {
	return fnGorm.GetDBP(ctx).Transaction(func(tx *gorm.DB) (err error) {
		var now = time.Now()
		var request = &CouponUseRequest{
			Id:           typ.NewUUID(),
			AccountId:    i.AccountId,
			UsedPoint:    i.Point,
			UsedCurrency: decimal.Zero,
			CreatedAt:    now,
		}

		var rows *gorm.DB
		if rows = tx.Create(request); rows.Error != nil {
			err = rows.Error
			return
		}

		var coupons Coupons
		if coupons, err = FindAllCoupons(tx, &FindCouponArgs{
			AccountId:  []typ.UUID{i.AccountId},
			HasBalance: fnReflect.Pointer(true),
			Lock:       fnReflect.Pointer(true),
		}); err != nil {
			return
		}

		if coupons.RestPoint().LessThan(i.Point) {
			err = ErrShortagePoint
			return
		}

		var restPoint = i.Point.Copy()
		for i := range coupons {
			var coupon = coupons[i]
			var prevBalance = coupon.CouponBalance

			var nextCouponBalance = &CouponBalance{
				Id:           typ.NewUUID(),
				CouponId:     coupon.Id,
				PrevPoint:    prevBalance.Point,
				PrevCurrency: prevBalance.Currency,
				CreatedAt:    now,
			}

			if coupon.CouponBalance.Point.GreaterThanOrEqual(restPoint) {
				nextCouponBalance.ChangedPoint = restPoint.Neg()
				nextCouponBalance.Point = prevBalance.Point.Add(nextCouponBalance.ChangedPoint)

				nextCouponBalance.ChangedCurrency = restPoint.Mul(coupon.Price).Neg()
				nextCouponBalance.Currency = prevBalance.Currency.Add(nextCouponBalance.ChangedCurrency)

				request.UsedCurrency = request.UsedCurrency.Add(restPoint.Mul(coupon.Price))
				restPoint = decimal.Zero
			} else {
				nextCouponBalance.Point = decimal.Zero
				nextCouponBalance.ChangedPoint = prevBalance.Point.Neg()

				nextCouponBalance.Currency = decimal.Zero
				nextCouponBalance.ChangedCurrency = prevBalance.Currency.Neg()
				request.UsedCurrency = request.UsedCurrency.Add(coupon.CouponBalance.Point)

				restPoint = restPoint.Sub(coupon.CouponBalance.Point)
			}

			if rows = tx.
				Model(&CouponBalance{}).
				Create(nextCouponBalance); rows.Error != nil {
				err = rows.Error
				return
			}

			var updates = map[string]any{
				"has_balance":       restPoint.GreaterThan(decimal.Zero),
				"coupon_balance_id": nextCouponBalance.Id,
				"updated_at":        now,
			}

			if rows = tx.
				Model(&Coupon{}).
				Where("`coupons`.`id` = ?", coupon.Id).
				Updates(updates); rows.RowsAffected == 0 {
				err = fmt.Errorf("not found coupon: couponId=%s", coupon.Id)
				return
			}

			var receipt = &CouponUseReceipt{
				Id:                 typ.NewUUID(),
				CouponUseRequestId: request.Id,
				CouponBalanceId:    nextCouponBalance.Id,
			}

			if rows = tx.Create(receipt); rows.Error != nil {
				err = rows.Error
				return
			}

			request.CouponUseReceipts = append(request.CouponUseReceipts, receipt)

			if restPoint.Equal(decimal.Zero) {
				break
			}
		}

		if rows = tx.
			Model(&CouponUseRequest{}).
			Where("`coupon_use_requests`.`id` = ?", request.Id).
			Updates(map[string]any{
				"`coupon_use_requests`.`used_currency`": request.UsedCurrency,
			}); rows.Error != nil {
			err = rows.Error
			return
		}

		return i.Fn(ctx, request)
	})
}
