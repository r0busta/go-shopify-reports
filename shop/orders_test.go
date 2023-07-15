package shop

import (
	"log"
	"testing"
	"time"

	"github.com/r0busta/go-shopify-graphql-model/v3/graph/model"
	"github.com/r0busta/go-shopify-reports/utils"
	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
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

func newFloat64(v float64) *float64 {
	return &v
}

func TestCalcTotalTurnover(t *testing.T) {
	from, to, err := utils.ParsePeriod([]string{"2020-04-01", "2020-04-01"})
	if err != nil {
		log.Fatalf("error parsing period: %s", err)
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
			name: "Single order with a sale transaction",
			args: args{
				from: *from,
				to:   *to,
				orders: []*model.Order{
					{
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
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
			name: "Two orders. One with a refund and another one with a transaction outside the period",
			args: args{
				from: *from,
				to:   *to,
				orders: []*model.Order{
					{
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.99"),
									},
								},
							},
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindRefund,
								Status:      model.OrderTransactionStatusSuccess,
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
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.02"),
									},
								},
							},
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 2, 11, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
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
			got, gotErr := CalcTotalTurnover(tt.args.orders, tt.args.from, tt.args.to)
			if gotErr != nil {
				t.Errorf("CalcTotalTurnover(), gotErr=%v, want %v", gotErr.Error(), nil)
			}
			if !got.Equal(*tt.want) {
				t.Errorf("CalcTotalTurnover() = %v, want %v", got.String(), tt.want.String())
			}
		})
	}
}

func TestCalcTotalNetTurnover(t *testing.T) {
	from, to, err := utils.ParsePeriod([]string{"2020-04-01", "2020-04-01"})
	if err != nil {
		log.Fatalf("error parsing period: %s", err)
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
				from:   *from,
				to:     *to,
				orders: []*model.Order{},
			},
			want: newDecimal(decimal.Zero),
		},
		{
			args: args{
				from: *from,
				to:   *to,
				orders: []*model.Order{
					{
						ShippingAddress: &model.MailingAddress{
							CountryCodeV2: newCountryCode(model.CountryCodeGb),
						},
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
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
				from: *from,
				to:   *to,
				orders: []*model.Order{
					{
						ShippingAddress: &model.MailingAddress{
							CountryCodeV2: newCountryCode(model.CountryCodeGb),
						},
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.01"),
									},
								},
							},
						},
						TaxLines: []model.TaxLine{
							{
								Rate: nil,
							},
						},
					},
				},
			},
			want: newDecimal(decimal.RequireFromString("10.01")),
		},
		{
			args: args{
				from: *from,
				to:   *to,
				orders: []*model.Order{
					{
						ShippingAddress: &model.MailingAddress{
							CountryCodeV2: newCountryCode(model.CountryCodeGb),
						},
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.01"),
									},
								},
							},
						},
						TaxLines: []model.TaxLine{
							{
								Rate: newFloat64(0.2),
							},
						},
					},
				},
			},
			want: newDecimal(decimal.RequireFromString("8.34")),
		},
		{
			args: args{
				from: *from,
				to:   *to,
				orders: []*model.Order{
					{
						ShippingAddress: &model.MailingAddress{
							CountryCodeV2: newCountryCode(model.CountryCodeGb),
						},
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.01"),
									},
								},
							},
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindRefund,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("5.01"),
									},
								},
							},
						},
						TaxLines: []model.TaxLine{
							{
								Rate: newFloat64(0.2),
							},
						},
					},
				},
			},
			want: newDecimal(decimal.RequireFromString("4.17")),
		},
		{
			args: args{
				from: *from,
				to:   *to,
				orders: []*model.Order{
					{
						ShippingAddress: &model.MailingAddress{
							CountryCodeV2: newCountryCode(model.CountryCodeGb),
						},
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.01"),
									},
								},
							},
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindRefund,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.01"),
									},
								},
							},
						},
						TaxLines: []model.TaxLine{
							{
								Rate: newFloat64(0.2),
							},
						},
					},
				},
			},
			want: newDecimal(decimal.RequireFromString("0.00")),
		},
		{
			args: args{
				from: *from,
				to:   *to,
				orders: []*model.Order{
					{
						ShippingAddress: &model.MailingAddress{
							CountryCodeV2: newCountryCode(model.CountryCodeGb),
						},
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.01"),
									},
								},
							},
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindRefund,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.01"),
									},
								},
							},
						},
						TaxLines: []model.TaxLine{
							{
								Rate: newFloat64(0.2),
							},
						},
					},
					{
						ShippingAddress: &model.MailingAddress{
							CountryCodeV2: newCountryCode(model.CountryCodeGb),
						},
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.99"),
									},
								},
							},
						},
						TaxLines: []model.TaxLine{
							{
								Rate: newFloat64(0.2),
							},
						},
					},
				},
			},
			want: newDecimal(decimal.RequireFromString("9.16")),
		},
		{
			args: args{
				from: *from,
				to:   *to,
				orders: []*model.Order{
					{
						ShippingAddress: &model.MailingAddress{
							CountryCodeV2: newCountryCode(model.CountryCodeGb),
						},
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.01"),
									},
								},
							},
						},
						TaxLines: []model.TaxLine{
							{
								Rate: newFloat64(0.2),
							},
						},
					},
					{
						ShippingAddress: &model.MailingAddress{
							CountryCodeV2: newCountryCode(model.CountryCodeGb),
						},
						Transactions: []model.OrderTransaction{
							{
								ProcessedAt: model.NewString(time.Date(2020, 4, 1, 10, 30, 0, 0, time.UTC).Format(ISO8601Layout)),
								Kind:        model.OrderTransactionKindSale,
								Status:      model.OrderTransactionStatusSuccess,
								Test:        false,
								AmountSet: &model.MoneyBag{
									ShopMoney: &model.MoneyV2{
										Amount: null.StringFrom("10.99"),
									},
								},
							},
						},
						TaxLines: []model.TaxLine{
							{
								Rate: newFloat64(0.2),
							},
						},
					},
				},
			},
			want: newDecimal(decimal.RequireFromString("17.50")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := CalcTotalNetTurnover(tt.args.orders, tt.args.from, tt.args.to)
			if gotErr != nil {
				t.Errorf("CalcTotalNetTurnover(), gotErr=%v, want %v", gotErr.Error(), nil)
			}
			if !got.Equal(*tt.want) {
				t.Errorf("CalcTotalNetTurnover() = %v, want %v", got.String(), tt.want.String())
			}
		})
	}
}
