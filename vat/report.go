package vat

import (
	"fmt"

	"github.com/r0busta/go-shopify-reports/shop"
	"github.com/r0busta/go-shopify-reports/utils"

	log "github.com/sirupsen/logrus"
)

type VATReturn interface {
	Report(period []string)
}

type FlatRateReturn struct {
}

func (s *FlatRateReturn) Report(period []string, useCached bool) {
	from, to, err := utils.ParsePeriod(period)
	if err != nil {
		log.Fatalln("error parsing period dates:", err)
	}

	shopClient := shop.NewClient()
	orders, err := shopClient.Order.ListCreatedBetween(*from, *to, useCached)
	if err != nil {
		log.Fatalf("error getting orders: %s", err)
	}
	log.Printf("Found %d orders", len(orders))

	totalTurnover, err := shop.CalcTotalTurnover(orders, *from, *to)
	if err != nil {
		log.Fatalf("Error calculating total turnover: %s", err)
	}

	fmt.Println("Total turnover, including VAT and EC sales (box 6):", totalTurnover.String())
}
