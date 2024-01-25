package mdPoint

import (
	"context"
	"errors"
	"github.com/d3v-friends/go-rdbms/fnGorm"
	"github.com/d3v-friends/go-snippet/fn/fnMatch"
	"github.com/d3v-friends/go-snippet/fn/fnReflect"
	"github.com/d3v-friends/go-snippet/typ"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

var (
	ErrNotFoundWallet      = errors.New("not found wallet")
	ErrShortageWalletPoint = errors.New("shortage wallet point")
)

type WalletUseRequest struct {
	Id        typ.UUID        `gorm:"primaryKey;type:char(36)"`
	UsedPoint decimal.Decimal `gorm:"type:decimal(20,2)"`
	Msg       string          `gorm:"type:text"`
	CreatedAt time.Time       `gorm:"index"`

	// ref
	WalletUseReceipt []*WalletUseReceipt `gorm:"foreignKey:WalletUseRequestId;references:Id"`
}

func (x *WalletUseRequest) Migrate() []fnGorm.Migrate {
	return make([]fnGorm.Migrate, 0)
}

type WalletUseReceipt struct {
	Id                 typ.UUID `gorm:"primaryKey"`
	WalletUseRequestId typ.UUID `gorm:"index"`
	WalletBalanceId    typ.UUID `gorm:"index"`

	// ref
	WalletUseRequest *WalletUseRequest `gorm:"foreignKey:WalletUseRequestId;references:Id"`
	WalletBalance    *WalletBalance    `gorm:"foreignKey:WalletBalanceId;references:Id"`
}

func (x *WalletUseReceipt) Migrate() []fnGorm.Migrate {
	return make([]fnGorm.Migrate, 0)
}

/*------------------------------------------------------------------------------------------------*/

type UseWalletArgs struct {
	Request UseWalletRequests
	Msg     string
	Fn      func(ctx context.Context, request *WalletUseRequest) (err error)
}

type UseWalletRequest struct {
	WalletId typ.UUID
	UsePoint decimal.Decimal
}

type UseWalletRequests []*UseWalletRequest

func (x UseWalletRequests) WalletIds() (ls []typ.UUID) {
	ls = make([]typ.UUID, len(x))
	for i := range x {
		ls[i] = x[i].WalletId
	}
	return
}

func (x UseWalletRequests) TotalUsePoint() (total decimal.Decimal) {
	total = decimal.Zero
	for i := range x {
		total = total.Add(x[i].UsePoint)
	}
	return
}

func UseWallet(ctx context.Context, i *UseWalletArgs) error {
	return fnGorm.GetDBP(ctx).Transaction(func(tx *gorm.DB) (err error) {
		// 지갑 잠금
		var wallets Wallets
		var walletIds = i.Request.WalletIds()
		if wallets, err = FindAllWallets(tx, &FindWalletArgs{
			Id:   walletIds,
			Lock: fnReflect.Pointer(true),
		}); err != nil {
			return
		}

		var now = time.Now()
		var request = &WalletUseRequest{
			Id:        typ.NewUUID(),
			UsedPoint: i.Request.TotalUsePoint(),
			Msg:       i.Msg,
			CreatedAt: now,
		}

		var rows *gorm.DB
		if rows = tx.
			Model(&WalletUseRequest{}).
			Omit("WalletUseReceipt").
			Create(request); rows.Error != nil {
			err = rows.Error
			return
		}

		request.WalletUseReceipt = make([]*WalletUseReceipt, 0)
		for idx := range i.Request {
			var req = i.Request[idx]
			var wallet *Wallet
			if wallet, err = fnMatch.Get(wallets, func(v *Wallet) bool {
				return v.Id == req.WalletId
			}); err != nil {
				err = ErrNotFoundWallet
				return
			}

			if wallet.
				WalletBalance.
				Point.
				Add(req.UsePoint).
				LessThan(decimal.Zero) {
				err = ErrShortageWalletPoint
				return
			}
			var prevWalletBalance = wallet.WalletBalance
			var nextWalletBalance = &WalletBalance{
				Id:           typ.NewUUID(),
				WalletId:     wallet.Id,
				Point:        prevWalletBalance.Point.Add(req.UsePoint),
				PrevPoint:    prevWalletBalance.Point,
				ChangedPoint: req.UsePoint,
				CreatedAt:    now,
			}

			if rows = tx.
				Model(&WalletBalance{}).
				Create(nextWalletBalance); rows.Error != nil {
				err = rows.Error
				return
			}

			if rows = tx.
				Model(&Wallet{}).
				Where("`wallets`.`id` = ?", wallet.Id).
				Updates(map[string]any{
					"`wallets`.`wallet_balance_id`": nextWalletBalance.Id,
					"`wallets`.`updated_at`":        now,
				}); rows.Error != nil {
				err = rows.Error
				return
			}

			var receipt = &WalletUseReceipt{
				Id:                 typ.NewUUID(),
				WalletUseRequestId: request.Id,
				WalletBalanceId:    nextWalletBalance.Id,
			}

			if rows = tx.
				Model(&WalletUseReceipt{}).
				Create(receipt); rows.Error != nil {
				err = rows.Error
				return
			}

			request.WalletUseReceipt = append(request.WalletUseReceipt, receipt)
		}

		if err = i.Fn(ctx, request); err != nil {
			return
		}

		return
	})
}
