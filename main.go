package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/r0busta/go-shopify-vat/shop"
	"github.com/r0busta/go-shopify-vat/vat"
)

const (
	datesRangeTimeLayout = "2006-01-02"
)

func newClients() (shopClient *shop.Client) {
	apiKey := os.Getenv("STORE_API_KEY")
	password := os.Getenv("STORE_PASSWORD")
	shopName := os.Getenv("STORE_NAME")
	if apiKey == "" || password == "" || shopName == "" {
		log.Panicln("Shopify app API Key and/or Password and/or Store Name not set")
	}

	shopClient = shop.NewClient(apiKey, password, shopName)

	return
}

func main() {
	from, err := time.Parse(datesRangeTimeLayout, "2020-07-01")
	if err != nil {
		log.Fatalln("error parsing time:", err)
	}
	to, err := time.Parse(datesRangeTimeLayout, "2020-07-31")
	if err != nil {
		log.Fatalln("error parsing time:", err)
	}

	shopClient := newClients()

	orders, err := shopClient.Order.ListCreatedBetween(&from, &to)
	if err != nil {
		log.Fatalln("error listing orders:", err)
	}

	log.Printf("Found %d orders", len(orders))
	// utils.WriteFormatedJSON(os.Stdout, orders)

	totalTurnover := vat.CalcTotalTurnover(orders, &from, &to)
	euSales := vat.CalcNetEUSales(orders, shop.CountryCode("GB"), &from, &to)

	fmt.Println("Total turnover, including VAT (box 6):", totalTurnover.String())
	fmt.Println("EU sales, excluding VAT (box 8):", euSales.String())
}
