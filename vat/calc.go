package vat

import (
	"fmt"
	"time"

	"github.com/r0busta/go-shopify-graphql-model/graph/model"
	"github.com/r0busta/go-shopify-reports/shop"
	"github.com/shopspring/decimal"
)

const vatRate = 0.2

func SumTransactions(transactions []*model.OrderTransaction, fromMin, toMax *time.Time) (*decimal.Decimal, error) {
	var total decimal.Decimal
	for _, t := range transactions {
		if t.Test {
			continue
		}

		processedAt, err := time.Parse(shop.ISO8601Layout, t.ProcessedAt.String)
		if err != nil {
			return nil, fmt.Errorf("error parsing processed at time: %s", err)
		}
		if processedAt.After(*fromMin) && processedAt.Before(*toMax) || processedAt.Equal(*fromMin) || processedAt.Equal(*toMax) {
			amount, err := shop.GetTransactionAmount(t)
			if err != nil {
				return nil, fmt.Errorf("error getting transaction amount: %s", err)
			}
			total = total.Add(*amount)
		}
	}
	return &total, nil
}

func CalcTotalTurnover(orders []*model.Order, from, to *time.Time) (*decimal.Decimal, error) {
	fromMin := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toMax := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 1e9-1, time.UTC)

	var total decimal.Decimal

	for _, o := range orders {
		sum, err := SumTransactions(o.Transactions, &fromMin, &toMax)
		if err != nil {
			return nil, fmt.Errorf("error getting transactions total: %s", err)
		}
		total = total.Add(*sum)
	}

	return &total, nil
}
