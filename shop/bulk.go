package shop

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/r0busta/go-shopify-uk-vat/fileutils"
	"github.com/r0busta/go-shopify-uk-vat/rand"
	"github.com/shurcooL/graphql"
)

type QueryCurrentBulkOperation struct {
	CurrentBulkOperation CurrentBulkOperation
}

type CurrentBulkOperation struct {
	ID             graphql.ID
	Status         graphql.String
	ErrorCode      graphql.String
	CreatedAt      graphql.String
	CompletedAt    graphql.String
	ObjectCount    graphql.String
	FileSize       graphql.String
	URL            graphql.String
	PartialDataURL graphql.String
}

type BulkOperationRunQuery struct {
	BulkOperation struct {
		ID graphql.ID
	}
	UserErrors []struct {
		Field   []graphql.String
		Message graphql.String
	}
}

type MutationBulkOperation struct {
	BulkOperationRunQuery BulkOperationRunQuery `graphql:"bulkOperationRunQuery(query: $query)"`
}

type BulkOperationCancel struct {
	BulkOperation struct {
		ID graphql.ID
	}
	UserErrors []struct {
		Field   []graphql.String
		Message graphql.String
	}
}

type MutationBulkOperationCancel struct {
	BulkOperationCancel BulkOperationCancel `graphql:"bulkOperationCancel(id: $id)"`
}

func postBulkQuery(shopifyGQL *graphql.Client, query string) error {
	m := MutationBulkOperation{}
	vars := map[string]interface{}{
		"query": graphql.String(query),
	}

	err := shopifyGQL.Mutate(context.Background(), &m, vars)
	if err != nil {
		return err
	}
	if len(m.BulkOperationRunQuery.UserErrors) > 0 {
		return fmt.Errorf("%+v", m.BulkOperationRunQuery.UserErrors)
	}

	return nil
}

func getBulkQueryResult(shopifyGQL *graphql.Client) (url string, err error) {
	q := QueryCurrentBulkOperation{}
	err = shopifyGQL.Query(context.Background(), &q, nil)
	if err != nil {
		return
	}

	// Start polling the operation's status
	for q.CurrentBulkOperation.Status == "CREATED" || q.CurrentBulkOperation.Status == "RUNNING" {
		log.Println("Bulk operation still running...")
		time.Sleep(1 * time.Second)

		err = shopifyGQL.Query(context.Background(), &q, nil)
		if err != nil {
			log.Printf("%+v", q)
			return
		}
	}
	log.Printf("Bulk operation finished with the status: %s", q.CurrentBulkOperation.Status)

	if q.CurrentBulkOperation.ErrorCode != "" {
		log.Printf("%+v", q)
		err = fmt.Errorf("Bulk operation error: %s", q.CurrentBulkOperation.ErrorCode)
		return
	}

	if q.CurrentBulkOperation.ObjectCount == "0" {
		err = fmt.Errorf("no results")
		return
	}

	url = string(q.CurrentBulkOperation.URL)
	return
}

func cancelRunningBulkQuery(shopifyGQL *graphql.Client) (err error) {
	q := QueryCurrentBulkOperation{}

	err = shopifyGQL.Query(context.Background(), &q, nil)
	if err != nil {
		return
	}

	if q.CurrentBulkOperation.Status == "RUNNING" {
		log.Println("Canceling running operation")
		operationID := q.CurrentBulkOperation.ID

		m := MutationBulkOperationCancel{}
		vars := map[string]interface{}{
			"id": graphql.ID(operationID),
		}

		err = shopifyGQL.Mutate(context.Background(), &m, vars)
		if err != nil {
			return err
		}
		if len(m.BulkOperationCancel.UserErrors) > 0 {
			return fmt.Errorf("%+v", m.BulkOperationCancel.UserErrors)
		}

		err = shopifyGQL.Query(context.Background(), &q, nil)
		if err != nil {
			return
		}

		for q.CurrentBulkOperation.Status == "CANCELING" {
			log.Println("Bulk operation still canceling...")
			err = shopifyGQL.Query(context.Background(), &q, nil)
			if err != nil {
				return
			}
		}
		log.Printf("Bulk operation cancelled")
	}

	return
}

func parseBulkQueryResult(resultFile string, out interface{}) (err error) {
	if reflect.TypeOf(out).Kind() != reflect.Ptr {
		err = fmt.Errorf("'records' is not a pointer")
		return
	}

	outValue := reflect.ValueOf(out)
	outSlice := outValue.Elem()
	if outSlice.Kind() != reflect.Slice {
		err = fmt.Errorf("'records' is not a  pointer to a slice interface")
		return
	}

	sliceItemType := outSlice.Type().Elem() // slice item type
	sliceItemKind := sliceItemType.Kind()
	itemType := sliceItemType // slice item underlying type
	if sliceItemKind == reflect.Ptr {
		itemType = itemType.Elem()
	}

	f, err := os.Open(resultFile)
	if err != nil {
		return
	}
	defer fileutils.CloseFile(f)

	reader := bufio.NewReader(f)
	json := jsoniter.ConfigFastest

	for {
		var line []byte
		line, err = reader.ReadBytes('\n')
		if err != nil {
			break
		}

		itemVal := reflect.New(itemType)
		err = json.Unmarshal(line, itemVal.Interface())
		if err != nil {
			return
		}

		if sliceItemKind == reflect.Ptr {
			outSlice.Set(reflect.Append(outSlice, itemVal))
		} else {
			outSlice.Set(reflect.Append(outSlice, itemVal.Elem()))
		}
	}

	if err != nil && err != io.EOF {
		return
	}

	err = nil
	return
}

func bulkQuery(shopify *graphql.Client, query string, out interface{}) (err error) {
	err = cancelRunningBulkQuery(shopify)
	if err != nil {
		return
	}

	err = postBulkQuery(shopify, query)
	if err != nil {
		return
	}

	url, err := getBulkQueryResult(shopify)
	if err != nil {
		return
	}

	filename := fmt.Sprintf("%s%s", rand.String(10), ".jsonl")
	resultFile := filepath.Join(os.TempDir(), filename)
	err = fileutils.DownloadFile(resultFile, url)
	if err != nil {
		return
	}

	err = parseBulkQueryResult(resultFile, out)
	if err != nil {
		return
	}

	return
}
