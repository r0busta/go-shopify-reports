package sales

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/r0busta/go-shopify-graphql-model/v3/graph/model"
	"github.com/r0busta/go-shopify-reports/shop"
	"github.com/r0busta/go-shopify-reports/utils"

	"github.com/shopspring/decimal"
	"github.com/thoas/go-funk"
	"github.com/tomlazar/table"
)

func ByTag(onlyTags []string, period []string, useCached bool, exportPath string) {
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

	type Stat struct {
		OrdersCount       int
		FulfilledQuantity int
		Revenue           decimal.Decimal
		Refunded          decimal.Decimal
	}

	stats := map[string]Stat{}

	for _, o := range orders {
		allTags := getTags(o.LineItems.Edges)
		for _, tag := range allTags {
			if len(onlyTags) > 0 && !funk.ContainsString(onlyTags, tag) {
				continue
			}

			if _, ok := stats[tag]; !ok {
				stats[tag] = Stat{
					Revenue:  decimal.Zero,
					Refunded: decimal.Zero,
				}
			}
			stat, _ := stats[tag]
			stat.OrdersCount++
			stat.FulfilledQuantity += getLineItemFulfilledQuantityByTag(o.LineItems.Edges, tag)

			sum, err := shop.SumTransactions(o.Transactions, model.OrderTransactionKindSale, *from, *to)
			if err != nil {
				log.Fatalf("error getting revenue: %s", err)
			}
			stat.Revenue = stat.Revenue.Add(*sum)

			refunded, err := shop.SumTransactions(o.Transactions, model.OrderTransactionKindRefund, *from, *to)
			if err != nil {
				log.Fatalf("error getting refunds total: %s", err)
			}
			stat.Refunded = stat.Refunded.Add(*refunded)
			stats[tag] = stat
		}
	}

	headers := []string{"Tag", "Orders Count", "Items Fulfilled", "Revenue", "Refunded", "Refund Ratio"}
	tab := table.Table{
		Headers: headers,
		Rows:    [][]string{},
	}
	for k, v := range stats {
		saleReturnRatio := decimal.Zero
		if !v.Revenue.IsZero() {
			saleReturnRatio = v.Refunded.Abs().Div(v.Revenue)
		}
		row := []string{
			k,
			strconv.Itoa(v.OrdersCount),
			strconv.Itoa(v.FulfilledQuantity),
			v.Revenue.StringFixed(2),
			fmt.Sprintf("(%s)", v.Refunded.Neg().StringFixed(2)),
			fmt.Sprintf("%s%%", saleReturnRatio.Mul(decimal.NewFromInt(100)).StringFixed(2)),
		}
		tab.Rows = append(tab.Rows, row)
	}
	tab.WriteTable(os.Stdout, nil)

	if exportPath != "" {
		out, err := os.Create(exportPath)
		if err != nil {
			log.Fatalln(err)
		}
		defer out.Close()

		w := csv.NewWriter(out)
		w.Write(headers)
		w.WriteAll(tab.Rows)
		if err := w.Error(); err != nil {
			log.Fatalln("error exporting csv:", err)
		}
	}
}

func ByVendor(period []string, useCached bool, exportPath string) {
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

	type Stat struct {
		OrdersCount       int
		FulfilledQuantity int
		Revenue           decimal.Decimal
		Refunded          decimal.Decimal
	}
	stats := map[string]Stat{}

	for _, o := range orders {
		vendors := getVendors(o.LineItems.Edges)
		for _, v := range vendors {
			if _, ok := stats[v]; !ok {
				stats[v] = Stat{
					Revenue:  decimal.Zero,
					Refunded: decimal.Zero,
				}
			}
			stat, _ := stats[v]
			stat.OrdersCount++
			stat.FulfilledQuantity += getLineItemFulfilledQuantityByVendor(o.LineItems.Edges, v)

			sum, err := shop.SumTransactions(o.Transactions, model.OrderTransactionKindSale, *from, *to)
			if err != nil {
				log.Fatalf("error getting revenue: %s", err)
			}
			stat.Revenue = stat.Revenue.Add(*sum)

			refunded, err := shop.SumTransactions(o.Transactions, model.OrderTransactionKindRefund, *from, *to)
			if err != nil {
				log.Fatalf("error getting refunds total: %s", err)
			}
			stat.Refunded = stat.Refunded.Add(*refunded)
			stats[v] = stat
		}
	}

	headers := []string{"Vendor", "Orders Count", "Items Fulfilled", "Revenue", "Refunded", "Refund Ratio"}
	tab := table.Table{
		Headers: headers,
		Rows:    [][]string{},
	}
	for k, v := range stats {
		saleReturnRatio := decimal.Zero
		if !v.Revenue.IsZero() {
			saleReturnRatio = v.Refunded.Abs().Div(v.Revenue)
		}
		row := []string{
			strings.Title(k),
			strconv.Itoa(v.OrdersCount),
			strconv.Itoa(v.FulfilledQuantity),
			v.Revenue.StringFixed(2),
			fmt.Sprintf("(%s)", v.Refunded.Neg().StringFixed(2)),
			fmt.Sprintf("%s%%", saleReturnRatio.Mul(decimal.NewFromInt(100)).StringFixed(2)),
		}
		tab.Rows = append(tab.Rows, row)
	}
	tab.WriteTable(os.Stdout, nil)

	if exportPath != "" {
		out, err := os.Create(exportPath)
		if err != nil {
			log.Fatalln(err)
		}
		defer out.Close()

		w := csv.NewWriter(out)
		w.Write(headers)
		w.WriteAll(tab.Rows)
		if err := w.Error(); err != nil {
			log.Fatalln("error exporting csv:", err)
		}
	}
}

func getTags(lineItems []model.LineItemEdge) []string {
	res := []string{}
	for _, li := range lineItems {
		res = append(res, getLineItemTags(li.Node)...)
	}
	return funk.UniqString(res)
}

func getLineItemTags(li *model.LineItem) []string {
	res := []string{}
	for _, t := range li.Product.Tags {
		res = append(res, strings.ToLower(t))
	}
	return res
}

func getLineItemFulfilledQuantityByTag(lineItems []model.LineItemEdge, tag string) int {
	res := 0
	for _, li := range lineItems {
		if funk.ContainsString(getLineItemTags(li.Node), tag) {
			res += li.Node.Quantity - li.Node.UnfulfilledQuantity
		}
	}
	return res
}

func getVendors(lineItems []model.LineItemEdge) []string {
	res := []string{}
	for _, li := range lineItems {
		res = append(res, getVendor(li.Node))
	}
	return funk.UniqString(res)
}

func getVendor(li *model.LineItem) string {
	if li.Vendor == nil {
		return ""
	}

	downcaseVendor := strings.ToLower(strings.TrimSpace(*li.Vendor))
	return downcaseVendor
}

func getLineItemFulfilledQuantityByVendor(lineItems []model.LineItemEdge, vendor string) int {
	res := 0
	for _, li := range lineItems {
		if getVendor(li.Node) == vendor {
			res += li.Node.Quantity - li.Node.UnfulfilledQuantity
		}
	}
	return res
}

func saleReturnRatioRange(ratio decimal.Decimal) int {
	d := ratio.Mul(decimal.NewFromFloat(100))
	v, _ := d.Sub(d.Mod(decimal.NewFromFloat(10))).Float64()
	if v > 100 {
		return 0
	}
	return 100 - int(v)
}
