package shop

import (
	diskstore "github.com/r0busta/go-object-store/disk"
	shopifygraphql "github.com/r0busta/go-shopify-graphql/v3"
)

const (
	ISO8601Layout = "2006-01-02T15:04:05Z"
)

type Client struct {
	shopifyClient *shopifygraphql.Client

	Order OrderService
}

func NewClient() *Client {
	c := &Client{
		shopifyClient: shopifygraphql.NewDefaultClient(),
	}

	c.Order = &OrderServiceOp{
		client: c,
		cache:  diskstore.New("_orders_cache.json"),
	}

	return c
}
