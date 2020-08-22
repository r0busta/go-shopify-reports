package shop

import (
	"github.com/shopspring/decimal"
	"github.com/shurcooL/graphql"
)

type MoneyBag struct {
	ShopMoney *MoneyV2 `json:"shopMoney,omitempty"`
}

type CurrencyCode graphql.String

type MoneyV2 struct {
	Amount       *decimal.Decimal `json:"amount,omitempty"`
	CurrencyCode CurrencyCode     `json:"currencyCode,omitempty"`
}
