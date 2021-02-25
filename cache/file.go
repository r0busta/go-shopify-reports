package cache

import (
	"encoding/json"
	"os"

	"github.com/r0busta/go-shopify-graphql-model/graph/model"
)

const cacheFilePath = ".cache.json"

func CheckCache() bool {
	_, err := os.Stat(cacheFilePath)
	if err != nil {
		return false
	}
	return true
}

func ReadCache() ([]*model.Order, error) {
	f, err := os.Open(cacheFilePath)
	if err != nil {
		return []*model.Order{}, err
	}
	defer f.Close()

	v := []*model.Order{}
	err = json.NewDecoder(f).Decode(&v)
	if err != nil {
		return []*model.Order{}, err
	}

	return v, nil
}

func WriteCache(orders []*model.Order) error {
	f, err := os.Create(cacheFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(orders)
	if err != nil {
		return err
	}

	return nil
}
