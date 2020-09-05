package vat

import (
	"github.com/r0busta/go-shopify-uk-vat/shop"
)

var euCountryCodes = []shop.CountryCode{
	"AT", "BE", "BG", "CY", "CZ", "DE", "DK", "EE", "ES", "FI", "FR", "GB", "GR", "HR", "HU", "IE", "IT", "LT", "LU", "LV", "MT", "NL", "PL", "PT", "RO", "SE", "SI", "SK",
}

func IsInEuropeanUnion(cc shop.CountryCode) bool {
	for _, c := range euCountryCodes {
		if c == cc {
			return true
		}
	}
	return false
}
