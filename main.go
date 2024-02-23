package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"thek3.com/recovery-tracking-app/model"
)

var c *dynamodb.Client

func main() {
	if c == nil {
		var err error
		c, err = newclient("local-dynodb-admin")
		if err != nil {
			log.Fatal(err)
		}
	}

	http.HandleFunc("GET /api/v1/records", handleGetRecord)
	http.HandleFunc("POST /api/v1/records", handlePostRecord)
	http.ListenAndServe(":8080", nil)
}

func handleGetRecord(w http.ResponseWriter, r *http.Request) {
	fmt.Println("im in")
	fmt.Fprintf(w, "print from insiders")
}

func handlePostRecord(w http.ResponseWriter, r *http.Request) {

	var record model.Record
	err := json.NewDecoder(r.Body).Decode(&record)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = putRecord(c, "Records", record)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}

func unsafeToAttrValue(in interface{}) types.AttributeValue {
	val, err := attributevalue.Marshal(in)
	if err != nil {
		log.Fatalf("could not marshal value `%v` with error: %v", in, err)
	}

	return val
}

func putRecord(c *dynamodb.Client, tableName string, record model.Record) (err error) {

	dynoItem := map[string]types.AttributeValue{
		"RecordId":   unsafeToAttrValue(record.RecordId),
		"CaseId":     unsafeToAttrValue(record.CaseId),
		"RecordName": unsafeToAttrValue(record.RecordName),
		"MetricUnit": unsafeToAttrValue(record.MetricUnit),
		"Details":    unsafeToAttrValue(record.Details),
	}
	_, err = c.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName), Item: dynoItem,
	})
	if err != nil {
		return err
	}

	return nil
}

func newclient(profile string) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("localhost"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:8000"}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "test", SecretAccessKey: "test", SessionToken: "",
				Source: "Mock credentials used above for local instance",
			},
		}),
	)
	if err != nil {
		return nil, err
	}

	c := dynamodb.NewFromConfig(cfg)
	return c, nil
}
