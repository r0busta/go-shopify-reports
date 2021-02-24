package cmd

import (
	"log"

	"github.com/r0busta/go-shopify-reports/vat"
)

type ReportCmd struct {
	Scheme string   `arg help:"Scheme (e.g. flat)" enum:"flat"`
	Period []string `arg required name:"date" help:"Period start and end dates (e.g. 2020-08-01 2020-10-31)"`
}

func (cmd *ReportCmd) Run(ctx *Globals) error {
	switch cmd.Scheme {
	case "flat":
		r := vat.FlatRateReturn{}
		r.Report(cmd.Period)
	default:
		log.Fatalln("Unimplemented scheme")
	}
	return nil
}
