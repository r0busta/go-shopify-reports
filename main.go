package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/r0busta/go-shopify-vat/shop"
	"github.com/r0busta/go-shopify-vat/vat"
)

var cli struct {
	Report struct {
		Period []string `arg required name:"date" help:"period start and end dates"`
	} `cmd help:"Print report for VAT refund purposes."`
}

const (
	periodFormatLayout = "2006-01-02"
)

func main() {
	ctx := kong.Parse(&cli)
	switch ctx.Command() {
	case "report <date>":
		report(cli.Report.Period)
	default:
		panic(ctx.Command())
	}
}

func report(period []string) {
	if len(period) != 2 {
		log.Fatalln("period is incorrect")
	}

	from, err := time.Parse(periodFormatLayout, period[0])
	if err != nil {
		log.Fatalln("error parsing time:", err)
	}
	to, err := time.Parse(periodFormatLayout, period[1])
	if err != nil {
		log.Fatalln("error parsing time:", err)
	}

	shopClient := newClients()

	orders, err := shopClient.Order.ListCreatedBetween(&from, &to)
	if err != nil {
		log.Fatalln("error listing orders:", err)
	}

	log.Printf("Found %d orders", len(orders))

	totalTurnover := vat.CalcTotalTurnover(orders, &from, &to)

	fmt.Println("Total turnover, including VAT (box 6):", totalTurnover.String())
}

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
