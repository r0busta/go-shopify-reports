package shop

import (
	shopifygraphql "github.com/r0busta/go-shopify-graphql"
	"github.com/shurcooL/graphql"
)

const (
	shopifyAPIVersion = "2020-10"
)

type Client struct {
	gql *graphql.Client

	Order OrderService
}

func newShopifyGraphQLClient(apiKey string, password string, shopName string) *graphql.Client {
	opts := []shopifygraphql.Option{
		shopifygraphql.WithVersion(shopifyAPIVersion),
		shopifygraphql.WithPrivateAppAuth(apiKey, password),
	}
	return shopifygraphql.NewClient(shopName, opts...)
}

func NewClient(apiKey string, password string, shopName string) *Client {
	c := &Client{gql: newShopifyGraphQLClient(apiKey, password, shopName)}
	c.Order = &OrderServiceOp{client: c}

	return c
}
