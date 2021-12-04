package shop

import (
	"fmt"
	"strings"
	"time"

	"github.com/r0busta/go-shopify-graphql-model/graph/model"
	"github.com/r0busta/go-shopify-reports/cache"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type OrderService interface {
	ListCreatedBetween(from, to time.Time, useCached bool) ([]*model.Order, error)
}

type OrderServiceOp struct {
	client *Client
}

var _ OrderService = &OrderServiceOp{}

func (s *OrderServiceOp) ListCreatedBetween(from, to time.Time, useCached bool) ([]*model.Order, error) {
	isCacheValid := false
	if useCached {
		isCacheValid = cache.CheckCache()
	}

	if useCached && isCacheValid {
		orders, err := cache.ReadCache()
		if err != nil {
			return []*model.Order{}, fmt.Errorf("error reading orders from cache: %s", err)
		}
		return orders, err
	}

	orders, err := s.listCreatedBetween(from, to)
	if err != nil {
		return []*model.Order{}, fmt.Errorf("error listing orders: %s", err)
	}
	err = cache.WriteCache(orders)
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
		d = d.Neg()
		return &d, nil
	default:
		return &decimal.Zero, nil
	}
}

func SumTransactions(transactions []*model.OrderTransaction, kind model.OrderTransactionKind, from, to time.Time) (*decimal.Decimal, error) {
	var total decimal.Decimal
	for _, t := range transactions {
		if t.Test {
			continue
		}

		processedAt, err := time.Parse(ISO8601Layout, t.ProcessedAt.String)
		if err != nil {
			return nil, fmt.Errorf("error parsing processed at time: %s", err)
		}
		if t.Kind == kind && (processedAt.After(from) && processedAt.Before(to) || processedAt.Equal(from) || processedAt.Equal(to)) {
			amount, err := GetTransactionAmount(t)
			if err != nil {
				return nil, fmt.Errorf("error getting transaction amount: %s", err)
			}
			total = total.Add(*amount)
		}
	}
	return &total, nil
}

func CalcTotalTurnover(orders []*model.Order, from, to time.Time) (*decimal.Decimal, error) {
	var total decimal.Decimal

	for _, o := range orders {
		revenue, err := SumTransactions(o.Transactions, model.OrderTransactionKindSale, from, to)
		if err != nil {
			return nil, fmt.Errorf("error getting transactions total: %s", err)
		}
		total = total.Add(*revenue)

		refunds, err := SumTransactions(o.Transactions, model.OrderTransactionKindRefund, from, to)
		if err != nil {
			return nil, fmt.Errorf("error getting transactions total: %s", err)
		}
		total = total.Add(*refunds)
	}

	return &total, nil
}

func GetOrderNetTotal(o *model.Order) (*decimal.Decimal, error) {
	total := decimal.Zero

	if o.TotalPriceSet != nil && o.TotalPriceSet.ShopMoney != nil && !o.TotalPriceSet.ShopMoney.Amount.IsZero() {
		revenue, err := decimal.NewFromString(o.TotalPriceSet.ShopMoney.Amount.ValueOrZero())
		if err != nil {
			return nil, fmt.Errorf("error parsing total price set: %s", err)
		}
		total = total.Add(revenue)
	}

	if o.TotalRefundedSet != nil && o.TotalRefundedSet.ShopMoney != nil && !o.TotalRefundedSet.ShopMoney.Amount.IsZero() {
		refunds, err := decimal.NewFromString(o.TotalRefundedSet.ShopMoney.Amount.ValueOrZero())
		if err != nil {
			return nil, fmt.Errorf("error parsing total refunded set: %s", err)
		}
		total = total.Sub(refunds)
	}

	for _, t := range o.TaxLines {
		if t.Rate == nil {
			continue
		}

		rate := decimal.NewFromFloat(*t.Rate)

		tax := total.Sub(total.Div(rate.Add(decimal.NewFromFloat(1)))).RoundBank(2)

		total = total.Sub(tax)
	}

	return &total, nil
}

func CalcTotalNetTurnover(orders []*model.Order, from, to time.Time) (*decimal.Decimal, error) {
	var total decimal.Decimal

	for _, o := range orders {
		orderNetTotal, err := GetOrderNetTotal(o)
		if err != nil {
			return nil, fmt.Errorf("calculating order net total: %s", err)
		}

		total = total.Add(*orderNetTotal)
	}

	return &total, nil
}
