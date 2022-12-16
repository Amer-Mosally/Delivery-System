package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type TableBasics struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}

func (basics TableBasics) Scan(date string) {
	var err error
	filtEx := expression.Name("date").Equal(expression.Value(date))
	expr, err := expression.NewBuilder().WithFilter(filtEx).Build()
	if err != nil {
		log.Printf("Couldn't build expressions for scan. Here's why: %v\n", err)
	} else {
		var response *dynamodb.ScanOutput
		response, err = basics.DynamoDbClient.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName:                 aws.String(basics.TableName),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
		})
		if err != nil {
			log.Printf("Couldn't scan for movies released between")
		} else {
			err = attributevalue.UnmarshalListOfMaps(response.Items, &movies)
			if err != nil {
				log.Printf("Couldn't unmarshal query response. Here's why: %v\n", err)
			}
		}
	}
}

func assign(w http.ResponseWriter, req *http.Request) {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	tableBasics := TableBasics{TableName: "courier_management",
		DynamoDbClient: dynamodb.NewFromConfig(sdkConfig)}

	tableBasics.Scan("2022/12/18")

	x := req.URL.Query()["x"][0]
	y := req.URL.Query()["y"][0]
	resp, _ := http.Get("http://localhost:9999?x=" + x + "&y=" + y)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Fprintf(w, string(body))
}

func calculate(w http.ResponseWriter, req *http.Request) {
	x := req.URL.Query()["x"][0]
	y := req.URL.Query()["y"][0]
	resp, _ := http.Get("http://localhost:9998?x=" + x + "&y=" + y)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Fprintf(w, string(body))
}

func main() {
	http.HandleFunc("/assign", assign)
	http.HandleFunc("/calculate", calculate)

	err := http.ListenAndServe("localhost:8888", nil)

	if err != nil {
		log.Fatal(err)
	}
}
