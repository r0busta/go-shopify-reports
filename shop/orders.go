package shop

import (
	"fmt"
	"strings"
	"time"

	diskstore "github.com/r0busta/go-object-store/disk"
	"github.com/r0busta/go-shopify-graphql-model/graph/model"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type OrderService interface {
	ListCreatedBetween(from, to time.Time, useCached bool) ([]*model.Order, error)
}

type OrderServiceOp struct {
	client *Client
	cache  *diskstore.Store
}

var _ OrderService = &OrderServiceOp{}

func (s *OrderServiceOp) ListCreatedBetween(from, to time.Time, useCached bool) ([]*model.Order, error) {
	isCacheValid := false
	if useCached {
		isCacheValid = s.cache.FileExists()
	}

	if useCached && isCacheValid {
		orders := []*model.Order{}
		err := s.cache.Read(&orders)
		if err != nil {
			return []*model.Order{}, fmt.Errorf("error reading orders from cache: %s", err)
		}
		return orders, err
	}

	orders, err := s.listCreatedBetween(from, to)
	if err != nil {
		return []*model.Order{}, fmt.Errorf("error listing orders: %s", err)
	}
	err = s.cache.Write(orders)
	if err != nil {
		return []*model.Order{}, fmt.Errorf("error caching orders: %s", err)
	}
	return orders, err
}

func (s *OrderServiceOp) listCreatedBetween(from, to time.Time) ([]*model.Order, error) {
	log.Printf("Getting orders in the range %s and %s", from.Format("Jan 2, 2006"), to.Format("Jan 2, 2006"))

	query := `
	{
		orders(query: "$query") {
			edges {
				node {
					id
					name
					createdAt
					totalPriceSet {
						shopMoney {
							amount
							currencyCode
						}
					}
					totalRefundedSet {
						shopMoney {
							amount
							currencyCode
						}
					}
					currentTotalTaxSet {
						shopMoney {
							amount
							currencyCode
						}
					}
					taxLines {
						rate
					}
					transactions {
						processedAt
						status
						kind
						test
						amountSet {
							shopMoney {
								amount
								currencyCode
							}
						}
					}
					shippingAddress{
						countryCodeV2
					}
					lineItems{
						edges{
							node{
								id
								product{
									id
									tags
								}
								vendor
								quantity
								unfulfilledQuantity
							}
						}
					}
				}
			}
		}
	}
	`
	query = strings.ReplaceAll(query, "$query", fmt.Sprintf(`
		(created_at:>='%[1]s' created_at:<='%[2]s')
		OR (updated_at:>='%[1]s' updated_at:<='%[2]s')
		OR (processed_at:>='%[1]s' processed_at:<='%[2]s')`, from.Format(ISO8601Layout), to.Format(ISO8601Layout)))
	var orders []*model.Order

	err := s.client.shopifyClient.BulkOperation.BulkQuery(query, &orders)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func GetTransactionAmount(t *model.OrderTransaction) (*decimal.Decimal, error) {
	if t.Status != model.OrderTransactionStatusSuccess {
		return &decimal.Zero, nil
	}

	switch t.Kind {
	case model.OrderTransactionKindSale:
		d, err := decimal.NewFromString(t.AmountSet.ShopMoney.Amount.String)
		if err != nil {
			return nil, fmt.Errorf("error: %s", err)
		}
		return &d, nil
	case model.OrderTransactionKindRefund:
		d, err := decimal.NewFromString(t.AmountSet.ShopMoney.Amount.String)
		if err != nil {
			return nil, fmt.Errorf("error: %s", err)
		}
		return &d, nil
	default:
		return &decimal.Zero, nil
	}
}

func SumTransactions(transactions []*model.OrderTransaction, kind model.OrderTransactionKind, from, to time.Time) (*decimal.Decimal, error) {
	var total decimal.Decimal

	for _, t := range transactions {
		if t.Test || t.Kind != kind {
			continue
		}

		processedAt, err := time.Parse(ISO8601Layout, t.ProcessedAt.String)
		if err != nil {
			return nil, fmt.Errorf("error parsing processed at time: %s", err)
		}
		if processedAt.After(from) && processedAt.Before(to) || processedAt.Equal(from) || processedAt.Equal(to) {
			amount, err := GetTransactionAmount(t)
			if err != nil {
				return nil, fmt.Errorf("error getting transaction amount: %s", err)
			}
			total = total.Add(*amount)
		}
	}

	return &total, nil
}

func CalcOrderTurnover(o *model.Order, from, to time.Time) (decimal.Decimal, error) {
	var income decimal.Decimal

	revenue, err := SumTransactions(o.Transactions, model.OrderTransactionKindSale, from, to)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error getting sales total: %s", err)
	}
	income = income.Add(*revenue)

	refunds, err := SumTransactions(o.Transactions, model.OrderTransactionKindRefund, from, to)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error getting refund total: %s", err)
	}
	income = income.Sub(*refunds)

	return income, nil
}

func CalcTotalTurnover(orders []*model.Order, from, to time.Time) (*decimal.Decimal, error) {
	var total decimal.Decimal

	for _, o := range orders {
		income, err := CalcOrderTurnover(o, from, to)
		if err != nil {
			return nil, fmt.Errorf("income calc: %s", err)
		}
		total = total.Add(income)
	}

	return &total, nil
}

func CalcOrderNetIncome(o *model.Order, from, to time.Time) (*decimal.Decimal, error) {
	income, err := CalcOrderTurnover(o, from, to)
	if err != nil {
		return nil, fmt.Errorf("income calc: %s", err)
	}

	if income.IsZero() {
		return &decimal.Zero, nil
	}

	if o.ShippingAddress != nil && o.ShippingAddress.CountryCodeV2 != nil && *o.ShippingAddress.CountryCodeV2 == model.CountryCodeGb {
		for _, t := range o.TaxLines {
			if t.Rate == nil {
				continue
			}

			invertRate := decimal.NewFromFloat(1 + *t.Rate)
			tax := income.Sub(income.Div(invertRate).Round(2))

			income = income.Sub(tax)
		}
	}

	return &income, nil
}

func GetOrderSaleTaxTotal(o *model.Order) (*decimal.Decimal, error) {
	res := decimal.Zero

	if o.CurrentTotalTaxSet != nil && o.CurrentTotalTaxSet.ShopMoney != nil && !o.CurrentTotalTaxSet.ShopMoney.Amount.IsZero() {
		tax, err := decimal.NewFromString(o.CurrentTotalTaxSet.ShopMoney.Amount.ValueOrZero())
		if err != nil {
			return nil, fmt.Errorf("error parsing total tax set: %s", err)
		}
		res = res.Add(tax)
	}

	return &res, nil
}

func CalcTotalNetTurnover(orders []*model.Order, from, to time.Time) (*decimal.Decimal, error) {
	var total decimal.Decimal

	for _, o := range orders {
		income, err := CalcOrderNetIncome(o, from, to)
		if err != nil {
			return nil, fmt.Errorf("calculating order net income: %s", err)
		}

		total = total.Add(*income)
	}

	return &total, nil
}

func CalcTotalSaleTax(orders []*model.Order, from, to time.Time) (*decimal.Decimal, error) {
	var tax decimal.Decimal

	for _, o := range orders {
		createdAt, err := time.Parse(ISO8601Layout, o.CreatedAt.String)
		if err != nil {
			return nil, fmt.Errorf("error parsing created at time: %s", err)
		}
		if (createdAt.Before(from) || createdAt.After(to)) && !createdAt.Equal(from) && !createdAt.Equal(to) {
			continue
		}

		orderTax, err := GetOrderSaleTaxTotal(o)
		if err != nil {
			return nil, fmt.Errorf("calculating order tax total: %s", err)
		}

		tax = tax.Add(*orderTax)
	}

	return &tax, nil
}
