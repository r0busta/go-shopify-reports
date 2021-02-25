package cmd

type Globals struct {
	Debug bool
}

type CLI struct {
	Globals

	VAT    VATReportCmd `cmd help:"Print report for VAT return purposes. Example: <cmd> vat flat 2020-05-01 2020-07-31"`
	Tag    TagCmd       `cmd help:"Print report by tag. Example: <cmd> jeans sales-by-tag 2020-05-01 2020-07-31"`
	Vendor VendorCmd    `cmd help:"Print report by vendor. Example: <cmd> sales-by-vendor 2020-05-01 2020-07-31"`
}
