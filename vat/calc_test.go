package vat

import (
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/r0busta/go-shopify-graphql-model/graph/model"
	"github.com/r0busta/go-shopify-reports/shop"
	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
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

func newCountryCode(v model.CountryCode) *model.CountryCode {
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
		orders []*model.Order
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
				orders: []*model.Order{
					{
						Transactions: []*model.OrderTransaction{
							{
								ProcessedAt: model.NewNullString(null.StringFrom(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(shop.ISO8601Layout))),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.01"),
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
				orders: []*model.Order{
					{
						Transactions: []*model.OrderTransaction{
							{
								ProcessedAt: model.NewNullString(null.StringFrom(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(shop.ISO8601Layout))),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.99"),
									},
								},
							},
							{
								ProcessedAt: model.NewNullString(null.StringFrom(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(shop.ISO8601Layout))),
								Kind:        "REFUND",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("5.01"),
									},
								},
							},
						},
					},
					{
						Transactions: []*model.OrderTransaction{
							{
								ProcessedAt: model.NewNullString(null.StringFrom(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(shop.ISO8601Layout))),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.02"),
									},
								},
							},
							{
								ProcessedAt: model.NewNullString(null.StringFrom(time.Date(2020, 4, 2, 11, 30, 0, 0, time.UTC).Format(shop.ISO8601Layout))),
								Kind:        "SALE",
								Status:      "SUCCESS",
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.02"),
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
			got, gotErr := CalcTotalTurnover(tt.args.orders, &tt.args.from, &tt.args.to)
			if gotErr != nil {
				t.Errorf("CalcTotalTurnover(), gotErr=%v, want %v", gotErr.Error(), nil)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalcTotalTurnover() = %v, want %v", got.String(), tt.want.String())
			}
		})
	}
}
