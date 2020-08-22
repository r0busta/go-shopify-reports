package vat

import (
	"time"

	"github.com/r0busta/go-shopify-vat/shop"
	"github.com/shopspring/decimal"
)

const vatRate = 0.2

func SumTransactions(transactions []shop.Transaction, fromMin, toMax *time.Time) *decimal.Decimal {
	var total decimal.Decimal
	for _, t := range transactions {
		if t.Test {
			continue
		}

		if t.ProcessedAt.After(*fromMin) && t.ProcessedAt.Before(*toMax) || t.ProcessedAt.Equal(*fromMin) || t.ProcessedAt.Equal(*toMax) {
			amount := shop.GetTransactionAmount(&t)
			total = total.Add(*amount)
		}
	}
	return &total
}

func CalcTotalTurnover(orders []*shop.Order, from, to *time.Time) *decimal.Decimal {
	fromMin := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toMax := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 1e9-1, time.UTC)

	var total decimal.Decimal

	for _, o := range orders {
		total = total.Add(*SumTransactions(o.Transactions, &fromMin, &toMax))
	}

	return &total
}

func CalcNetEUSales(orders []*shop.Order, shopLocation shop.CountryCode, from, to *time.Time) *decimal.Decimal {
	fromMin := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toMax := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 1e9-1, time.UTC)

	var total decimal.Decimal
	for _, o := range orders {
		if o.ShippingAddress.CountryCodeV2 == shopLocation || !IsInEuropeanUnion(o.ShippingAddress.CountryCodeV2) {
			continue
		}

		amount := SumTransactions(o.Transactions, &fromMin, &toMax)
		tax := amount.Mul(decimal.NewFromFloat(vatRate / (1 + vatRate))).Round(2)
		net := amount.Sub(tax)
		total = total.Add(net)
	}

	return &total
}
