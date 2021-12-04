package corporatetax

import (
	"fmt"
	"log"

	"github.com/r0busta/go-shopify-reports/shop"
	"github.com/r0busta/go-shopify-reports/utils"
)

func Report(period []string, useCached bool) {
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

	totalTurnover, err := shop.CalcTotalNetTurnover(orders, *from, *to)
	if err != nil {
		log.Fatalf("Error calculating total turnover: %s", err)
	}

	totalSaleTax, err := shop.CalcTotalSaleTax(orders, *from, *to)
	if err != nil {
		log.Fatalf("Error calculating total tax: %s", err)
	}

	fmt.Println("Total turnover (excl. VAT):", totalTurnover.String())
	fmt.Println("Total tax (VAT):", totalSaleTax.String())
}
