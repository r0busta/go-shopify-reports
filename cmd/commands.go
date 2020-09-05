package cmd

import (
	"log"

	"github.com/r0busta/go-shopify-uk-vat/vat"
)

type ReportCmd struct {
	Scheme string   `arg help:"Scheme" enum:"flat"`
	Period []string `arg required name:"date" help:"period start and end dates"`
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
