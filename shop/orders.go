package shop

import (
	"fmt"
	"strings"
	"time"

	"github.com/r0busta/go-shopify-graphql-model/graph/model"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const (
	queryTimeLayout = "2006-01-02T15:04:05Z"
	fromMinTime     = "00:00:00Z"
	toMaxTime       = "23:59:59Z"
)

type OrderService interface {
	ListCreatedBetween(from, to *time.Time) ([]*model.Order, error)
}

type OrderServiceOp struct {
	client *Client
}

var _ OrderService = &OrderServiceOp{}

func (s *OrderServiceOp) ListCreatedBetween(from, to *time.Time) ([]*model.Order, error) {
	fromMin := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toMax := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 1e9-1, time.UTC)
	log.Printf("Getting orders in the range %s-%s", fromMin, toMax)

	query := `
	{
		orders(query: "$query") {
			edges {
				node {
					name
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
				}
			}
		}
	}
	`
	query = strings.ReplaceAll(query, "$query", fmt.Sprintf(`
		(created_at:>='%[1]s' created_at:<='%[2]s')
		OR (updated_at:>='%[1]s' updated_at:<='%[2]s')
		OR (processed_at:>='%[1]s' processed_at:<='%[2]s')`, fromMin.Format(queryTimeLayout), toMax.Format(queryTimeLayout)))
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
