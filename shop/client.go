package shop

import (
	"log"
	"os"

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

func NewDefaultClients() (shopClient *Client) {
	apiKey := os.Getenv("STORE_API_KEY")
	password := os.Getenv("STORE_PASSWORD")
	shopName := os.Getenv("STORE_NAME")
	if apiKey == "" || password == "" || shopName == "" {
		log.Panicln("Shopify app API Key and/or Password and/or Store Name not set")
	}

	shopClient = NewClient(apiKey, password, shopName)

	return
}

func NewClient(apiKey string, password string, shopName string) *Client {
	c := &Client{gql: newShopifyGraphQLClient(apiKey, password, shopName)}
	c.Order = &OrderServiceOp{client: c}

	return c
}

func newShopifyGraphQLClient(apiKey string, password string, shopName string) *graphql.Client {
	opts := []shopifygraphql.Option{
		shopifygraphql.WithVersion(shopifyAPIVersion),
		shopifygraphql.WithPrivateAppAuth(apiKey, password),
	}
	return shopifygraphql.NewClient(shopName, opts...)
}
