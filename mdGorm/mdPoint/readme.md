# mdPoint

* 포인트 관리 모델

# ERD

## coupon

~~~mermaid
erDiagram
    Account {
        UUID id
        Time created_at
        Time updated_at
    }

    Coupon {
        UUID id
        Bool has_balance
        UUID account_id
        UUID coupon_balance_id
        Decimal currency
        Decimal price
        Decimal point
        Time created_at
        Time updated_at
    }

    CouponBalance {
        UUID id
        UUID coupon_id
        Decimal point
        Decimal prev_point
        Decimal changed_point
        Decimal currency
        Decimal prev_currency
        Decimal changed_currency
        Time created_at
    }

    CouponUseRequest {
        UUID id
        UUID account_id
        Decimal used_point
        Decimal used_currency
        Text msg
        Time created_at
    }

    CouponUseReceipt {
        UUID id
        UUID coupon_use_request_id
        UUID coupon_balance_id
    }

%% Coupon %%
    Account ||--o{ Coupon: has
    Coupon ||--|{ CouponBalance: has
    CouponUseRequest ||--|{ CouponUseReceipt: has
    CouponUseReceipt ||--|| CouponBalance: has


~~~

## wallet point

* wallet 을 lock 하여 트렌젝션하여 잔고 거래의 무결성을 지킨다

~~~mermaid
erDiagram
    Account {
        UUID id
        Time created_at
        Time updated_at
    }

    Wallet {
        UUID id
        UUID account_id
        UUID wallet_balance_id
        Time created_at
        Time updated_at
    }

    WalletBalance {
        UUID id
        UUID wallet_id
        Decimal point
        Decimal prev_point
        Decimal changed_point
        Text memo
        Time created_at
    }

    WalletUseRequest {
        UUID id
        Decimal used_point
        Text msg
        Time created_at
    }

    WalletUseReceipt {
        UUID id
        UUID wallet_use_request_id
        UUID wallet_balance_id
    }

    Account ||--o{ Wallet: has
    Wallet ||--|{ WalletBalance: has
    WalletUseRequest ||--|{ WalletUseReceipt: has
    WalletUseReceipt ||--|| WalletBalance: has
~~~

# simple view

~~~mysql
CREATE VIEW coupons_view AS
(
SELECT C.id,
       C.account_id,
       C.point,
       C.currency,
       U.point             as left_point,
       (U.point * C.price) AS left_balance,
       C.created_at
FROM coupons C,
     coupon_balances U
WHERE C.coupon_balance_id = U.id
ORDER BY C.created_at DESC
    );

~~~
