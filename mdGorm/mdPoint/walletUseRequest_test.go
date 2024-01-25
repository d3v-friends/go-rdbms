package mdPoint

import (
	"context"
	"errors"
	"github.com/d3v-friends/go-snippet/fn/fnPanic"
	"github.com/d3v-friends/go-snippet/typ"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sync"
	"testing"
)

func TestWalletUseRequest(test *testing.T) {
	var tools = NewTestTool()

	test.Run("usePoint", func(t *testing.T) {
		var account = tools.NewAccount()
		fnPanic.On(UseWallet(tools.Context(), &UseWalletArgs{
			Request: UseWalletRequests{
				{
					WalletId: account.Wallets[0].Id,
					UsePoint: decimal.NewFromInt(10000),
				},
			},
			Msg: "message",
			Fn: func(ctx context.Context, request *WalletUseRequest) (err error) {
				assert.Equal(t, "10000", request.UsedPoint.String())
				assert.Equal(t, "message", request.Msg)
				return
			},
		}))
	})

	test.Run("shortage", func(t *testing.T) {
		var account = tools.NewAccount()
		var err = UseWallet(tools.Context(), &UseWalletArgs{
			Request: UseWalletRequests{
				{
					WalletId: account.Wallets[0].Id,
					UsePoint: decimal.NewFromInt(10000).Neg(),
				},
			},
			Msg: "message",
			Fn: func(ctx context.Context, request *WalletUseRequest) (err error) {
				return
			},
		})

		assert.Error(t, err, ErrShortagePoint)
	})

	test.Run("on error rollback", func(t *testing.T) {
		var account = tools.NewAccount()
		var rollbackErr = errors.New("rollback")
		var err = UseWallet(tools.Context(), &UseWalletArgs{
			Request: UseWalletRequests{
				{
					WalletId: account.Wallets[0].Id,
					UsePoint: decimal.NewFromInt(10000),
				},
			},
			Msg: "message",
			Fn: func(ctx context.Context, request *WalletUseRequest) error {
				return rollbackErr
			},
		})

		assert.Error(t, err, rollbackErr)
	})

	test.Run("async trx", func(t *testing.T) {
		var account = tools.NewAccount()
		var try int64 = 10
		var chargePoint int64 = 10000
		var wg = &sync.WaitGroup{}

		var totalChargePoint = decimal.Zero
		wg.Add(int(try))
		for i := 0; i < int(try); i++ {
			var ch = decimal.NewFromInt(rand.Int63n(chargePoint))
			totalChargePoint = totalChargePoint.Add(ch)

			go func(w *sync.WaitGroup, c decimal.Decimal, tt *TestTool, wl *Wallet) {
				fnPanic.On(UseWallet(tt.Context(), &UseWalletArgs{
					Request: UseWalletRequests{
						{
							WalletId: wl.Id,
							UsePoint: c,
						},
					},
					Msg: "",
					Fn: func(ctx context.Context, request *WalletUseRequest) (err error) {
						return
					},
				}))
				w.Done()
			}(wg, ch, tools, account.Wallets[0])
		}

		wg.Wait()

		var loadWallet = fnPanic.Get(FindOneWalletCtx(tools.Context(), &FindWalletArgs{
			Id: []typ.UUID{
				account.Wallets[0].Id,
			},
		}))

		assert.Equal(t, totalChargePoint.String(), loadWallet.WalletBalance.Point.String())

	})

}
