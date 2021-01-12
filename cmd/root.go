package cmd

type Globals struct {
	Debug bool
}

type CLI struct {
	Globals

	Report ReportCmd `cmd help:"Print various sale numbers for VAT return purposes. Example: <cmd> report flat 2020-05-01 2020-07-31"`
}
