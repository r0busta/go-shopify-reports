package shop

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	queryTimeLayout = "2006-01-02T15:04:05Z"
	fromMinTime     = "00:00:00Z"
	toMaxTime       = "23:59:59Z"

	transactionStatusSuccess = "SUCCESS"

	transactionKindSale   = "SALE"
	transactionKindRefund = "REFUND"
)

type CountryCode string

type MailingAddress struct {
	CountryCodeV2 CountryCode `json:"countryCodeV2,omitempty"`
}

type TransactionStatus string

type TransactionKind string

type Transaction struct {
	ProcessedAt *time.Time        `json:"processedAt,omitempty"`
	Status      TransactionStatus `json:"status,omitempty"`
	Kind        TransactionKind   `json:"kind,omitempty"`
	Test        bool              `json:"test,omitempty"`
	AmountSet   *MoneyBag         `json:"amountSet,omitempty"`
}

type Order struct {
	Name            string          `json:"name,omitempty"`
	ShippingAddress *MailingAddress `json:"shippingAddress,omitempty"`
	Transactions    []Transaction   `json:"transactions,omitempty"`
}

type OrderService interface {
	ListCreatedBetween(from, to *time.Time) ([]*Order, error)
}

type OrderServiceOp struct {
	client *Client
}

var _ OrderService = &OrderServiceOp{}

func (s *OrderServiceOp) ListCreatedBetween(from, to *time.Time) ([]*Order, error) {
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
	var orders []*Order

	err := bulkQuery(s.client.gql, query, &orders)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func GetTransactionAmount(t *Transaction) *decimal.Decimal {
	if t.Status != transactionStatusSuccess {
		return &decimal.Zero
	}

	switch t.Kind {
	case transactionKindSale:
		return t.AmountSet.ShopMoney.Amount
	case transactionKindRefund:
		amount := t.AmountSet.ShopMoney.Amount.Neg()
		return &amount
	default:
		return &decimal.Zero
	}
}
