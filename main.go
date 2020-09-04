package main

import (
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/joho/godotenv"
	"github.com/r0busta/go-shopify-vat/cmd"
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
		kong.Description("Get various sale numbers for VAT return from Shopify store"),
		kong.UsageOnError())
	err = ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
