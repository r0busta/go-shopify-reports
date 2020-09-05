package main

import (
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/joho/godotenv"
	"github.com/r0busta/go-shopify-uk-vat/cmd"
)

func main() {
	env := os.Getenv("ENV")
	if "" == env {
		env = "dev"
	}

	err := godotenv.Load(".env." + env + ".local")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cli := cmd.CLI{}

	ctx := kong.Parse(&cli,
		kong.Name("vat"),
		kong.Description("Get various reports from Shopify store for filling a VAT return with Her Majesty's Revenue and Customs (aka HMRC)."),
		kong.UsageOnError())
	err = ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
