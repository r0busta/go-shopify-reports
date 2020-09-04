package vat

import (
	"fmt"
	"log"
	"time"

	"github.com/r0busta/go-shopify-vat/shop"
)

type VATReturn interface {
	Report(period []string)
}

type FlatRateReturn struct {
}

const (
	periodFormatLayout = "2006-01-02"
)

func (s *FlatRateReturn) Report(period []string) {
	from, to, err := parsePeriod(period)
	if err != nil {
		log.Fatalln("error parsing period dates:", err)
	}

	shopClient := shop.NewDefaultClients()

	orders, err := shopClient.Order.ListCreatedBetween(from, to)
	if err != nil {
		log.Fatalln("error listing orders:", err)
	}

	log.Printf("Found %d orders", len(orders))

	totalTurnover := CalcTotalTurnover(orders, from, to)

	fmt.Println("Total turnover, including VAT and EC sales (box 6):", totalTurnover.String())
}

func parsePeriod(period []string) (*time.Time, *time.Time, error) {
	if len(period) != 2 {
		return nil, nil, fmt.Errorf("expected `from` and `to` period dates")
	}

	from, err := time.Parse(periodFormatLayout, period[0])
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing `from` date: %s", err)
	}

	to, err := time.Parse(periodFormatLayout, period[1])
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing `to` date: %s", err)
	}

	return &from, &to, nil
}
