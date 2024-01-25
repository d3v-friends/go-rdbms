package mdPoint

import (
	"github.com/d3v-friends/go-rdbms/fnGorm"
)

var MigrateModels = []fnGorm.MigrateModel{
	&Account{},
	&Coupon{},
	&CouponBalance{},
	&CouponUseRequest{},
	&CouponUseReceipt{},
	&Wallet{},
	&WalletBalance{},
	&WalletUseRequest{},
	&WalletUseReceipt{},
}
