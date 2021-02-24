package shop

import (
	shopifygraphql "github.com/r0busta/go-shopify-graphql/v3"
)

const (
	shopifyAPIVersion = "2020-10"

	ISO8601Layout = "2006-01-02T15:04:05Z"
)

type Client struct {
	shopifyClient *shopifygraphql.Client

	Order OrderService
}

func NewClient() *Client {
	c := &Client{shopifyClient: shopifygraphql.NewDefaultClient()}
	c.Order = &OrderServiceOp{client: c}

	return c
}
