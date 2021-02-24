package vat

import (
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/r0busta/go-shopify-reports/shop"
	"github.com/shopspring/decimal"
)

const (
	datesRangeTimeLayout = "2006-01-02"
)

func newDecimal(v decimal.Decimal) *decimal.Decimal {
	return &v
}

func newTime(v time.Time) *time.Time {
	return &v
}

func TestCalcTotalTurnover(t *testing.T) {
	from, err := time.Parse(datesRangeTimeLayout, "2020-04-01")
	if err != nil {
		log.Fatalln("error parsing time:", err)
	}
	to, err := time.Parse(datesRangeTimeLayout, "2020-04-01")
	if err != nil {
		log.Fatalln("error parsing time:", err)
	}

	type args struct {
		orders []*shop.Order
		from   time.Time
		to     time.Time
	}
	tests := []struct {
		name string
		args args
		want *decimal.Decimal
	}{
		{
			args: args{
				from: from,
				to:   to,
				orders: []*shop.Order{
					{
						Transactions: []shop.Transaction{
							{
								ProcessedAt: newTime(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC)),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &shop.MoneyBag{
									ShopMoney: &shop.MoneyV2{
										Amount: newDecimal(decimal.RequireFromString("10.01")),
									},
								},
							},
						},
					},
				},
			},
			want: newDecimal(decimal.RequireFromString("10.01")),
		},
		{
			args: args{
				from: from,
				to:   to,
				orders: []*shop.Order{
					{
						Transactions: []shop.Transaction{
							{
								ProcessedAt: newTime(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC)),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &shop.MoneyBag{
									ShopMoney: &shop.MoneyV2{
										Amount: newDecimal(decimal.RequireFromString("10.99")),
									},
								},
							},
							{
								ProcessedAt: newTime(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC)),
								Kind:        "REFUND",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &shop.MoneyBag{
									ShopMoney: &shop.MoneyV2{
										Amount: newDecimal(decimal.RequireFromString("5.01")),
									},
								},
							},
						},
					},
					{
						Transactions: []shop.Transaction{
							{
								ProcessedAt: newTime(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC)),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &shop.MoneyBag{
									ShopMoney: &shop.MoneyV2{
										Amount: newDecimal(decimal.RequireFromString("10.02")),
									},
								},
							},
							{
								ProcessedAt: newTime(time.Date(2020, 4, 2, 11, 30, 0, 0, time.UTC)),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &shop.MoneyBag{
									ShopMoney: &shop.MoneyV2{
										Amount: newDecimal(decimal.RequireFromString("10.02")),
									},
								},
							},
						},
					},
				},
			},
			want: newDecimal(decimal.RequireFromString("16.00")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcTotalTurnover(tt.args.orders, &tt.args.from, &tt.args.to)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalcTotalTurnover() = %v, want %v", got.String(), tt.want.String())
			}
		})
	}
}

func TestCalcNetEUSales(t *testing.T) {
	from, err := time.Parse(datesRangeTimeLayout, "2020-04-01")
	if err != nil {
		log.Fatalln("error parsing time:", err)
	}
	to, err := time.Parse(datesRangeTimeLayout, "2020-04-01")
	if err != nil {
		log.Fatalln("error parsing time:", err)
	}

	type args struct {
		orders       []*shop.Order
		shopLocation shop.CountryCode
		from         time.Time
		to           time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    *decimal.Decimal
		wantErr bool
	}{
		{
			args: args{
				from:         from,
				to:           to,
				shopLocation: shop.CountryCode("GB"),
				orders: []*shop.Order{
					{
						Transactions: []shop.Transaction{
							{
								ProcessedAt: newTime(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC)),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &shop.MoneyBag{
									ShopMoney: &shop.MoneyV2{
										Amount: newDecimal(decimal.RequireFromString("10.02")),
									},
								},
							},
						},
						ShippingAddress: &shop.MailingAddress{
							CountryCodeV2: shop.CountryCode("GB"),
						},
					},
					{
						Transactions: []shop.Transaction{
							{
								ProcessedAt: newTime(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC)),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &shop.MoneyBag{
									ShopMoney: &shop.MoneyV2{
										Amount: newDecimal(decimal.RequireFromString("10.02")),
									},
								},
							},
							{
								ProcessedAt: newTime(time.Date(2020, 4, 1, 12, 30, 0, 0, time.UTC)),
								Kind:        "REFUND",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &shop.MoneyBag{
									ShopMoney: &shop.MoneyV2{
										Amount: newDecimal(decimal.RequireFromString("5.01")),
									},
								},
							},
						},
						ShippingAddress: &shop.MailingAddress{
							CountryCodeV2: shop.CountryCode("DE"),
						},
					},
					{
						Transactions: []shop.Transaction{
							{
								ProcessedAt: newTime(time.Date(2020, 4, 2, 11, 30, 0, 0, time.UTC)),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &shop.MoneyBag{
									ShopMoney: &shop.MoneyV2{
										Amount: newDecimal(decimal.RequireFromString("20.02")),
									},
								},
							},
							{
								ProcessedAt: newTime(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC)),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &shop.MoneyBag{
									ShopMoney: &shop.MoneyV2{
										Amount: newDecimal(decimal.RequireFromString("10.02")),
									},
								},
							},
						},
						ShippingAddress: &shop.MailingAddress{
							CountryCodeV2: shop.CountryCode("SE"),
						},
					},
				},
			},
			want: newDecimal(decimal.RequireFromString("12.53")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcNetEUSales(tt.args.orders, tt.args.shopLocation, &tt.args.from, &tt.args.to)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalcNetEUSales() = %v, want %v", got.String(), tt.want.String())
			}
		})
	}
}
