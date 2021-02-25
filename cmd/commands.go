package cmd

import (
	"github.com/r0busta/go-shopify-reports/sales"
	"github.com/r0busta/go-shopify-reports/vat"
	log "github.com/sirupsen/logrus"
)

type VATReportCmd struct {
	Scheme string   `arg help:"Scheme (e.g. flat)" enum:"flat"`
	Period []string `arg required name:"date" help:"Period start and end dates (e.g. 2020-08-01 2020-10-31)"`
	Cached bool     `name:"cached" help:"Use cached results"`
}

type TagCmd struct {
	Tags       []string `required name:"tags" help:"Tags to report sales for"`
	Period     []string `arg required name:"date" help:"Period start and end dates (e.g. 2020-08-01 2020-10-31)"`
	Cached     bool     `name:"cached" help:"Use cached results"`
	ExportPath string   `name:"out" help:"Define the path to export results as a CSV file"`
}

type VendorCmd struct {
	Period     []string `arg required name:"date" help:"Period start and end dates (e.g. 2020-08-01 2020-10-31)"`
	Cached     bool     `name:"cached" help:"Use cached results"`
	ExportPath string   `name:"out" help:"Define the path to export results as a CSV file"`
}

func (cmd *VATReportCmd) Run(ctx *Globals) error {
	switch cmd.Scheme {
	case "flat":
		r := vat.FlatRateReturn{}
		r.Report(cmd.Period, cmd.Cached)
	default:
		log.Fatalln("Unimplemented scheme")
	}
	return nil
}

func (cmd *TagCmd) Run(ctx *Globals) error {
	sales.ByTag(cmd.Tags, cmd.Period, cmd.Cached, cmd.ExportPath)
	return nil
}

func (cmd *VendorCmd) Run(ctx *Globals) error {
	sales.ByVendor(cmd.Period, cmd.Cached, cmd.ExportPath)
	return nil
}
