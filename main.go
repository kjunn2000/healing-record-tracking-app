package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"thek3.com/recovery-tracking-app/model"
)

var c *dynamodb.Client

type DynamoAttribute map[string]types.AttributeValue

func main() {
	if c == nil {
		var err error
		c, err = newclient("local-dynodb-admin")
		if err != nil {
			log.Fatal(err)
		}
	}

	http.HandleFunc("GET /api/v1/records/{recordId}", handleGetRecord)
	http.HandleFunc("POST /api/v1/records", handlePostRecord)
	http.ListenAndServe(":8080", nil)
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

func handleGetRecord(w http.ResponseWriter, r *http.Request) {
	recordId, _ := strconv.Atoi(r.PathValue("recordId"))
	dynoItem := DynamoAttribute{
		"RecordId": unsafeToAttrValue(recordId),
		// "CaseId": unsafeToAttrValue(r.PathValue("caseId")),
	}

	records, err := getRecords(c, "Records", dynoItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var record model.Record
	attributevalue.UnmarshalMap(records, &record)
	json.NewEncoder(w).Encode(record)
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
		print("Error here: ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}

func putRecord(c *dynamodb.Client, tableName string, record model.Record) (err error) {

	dynoItem := DynamoAttribute{
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

func getRecords(c *dynamodb.Client, tableName string, key DynamoAttribute) (item DynamoAttribute, err error) {
	resp, err := c.GetItem(context.TODO(), &dynamodb.GetItemInput{Key: key, TableName: aws.String(tableName)})
	if err != nil {
		return nil, err
	}

	return resp.Item, nil
}

func unsafeToAttrValue(in interface{}) types.AttributeValue {
	val, err := attributevalue.Marshal(in)
	if err != nil {
		log.Fatalf("could not marshal value `%v` with error: %v", in, err)
	}

	return val
}
