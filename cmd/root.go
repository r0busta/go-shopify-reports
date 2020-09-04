package cmd

type Globals struct {
	Debug bool
}

type CLI struct {
	Globals

	Report ReportCmd `cmd help:"Print various sale numbers for VAT return purposes."`
}
