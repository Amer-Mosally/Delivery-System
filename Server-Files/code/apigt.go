package main

import (
	"context"
	"fmt"
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

type Courier struct {
	ID        int
	name      string
	x         string
	y         string
	available bool
}

type Shipment struct {
	ID   int
	date string
	x    string
	y    string
}

func (basics TableBasics) ScanShipments(date string) []Shipment {
	var err error
	var shipments []Shipment
	filtEx := expression.Name("date").Equal(expression.Value(date))
	projEx := expression.NamesList(
		expression.Name("ID"), expression.Name("x"), expression.Name("y"))
	expr, err := expression.NewBuilder().WithFilter(filtEx).WithProjection(projEx).Build()
	if err != nil {
		log.Printf("Couldn't build expressions for scan. Here's why: %v\n", err)
	} else {
		var response *dynamodb.ScanOutput
		response, err = basics.DynamoDbClient.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName:                 aws.String(basics.TableName),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ProjectionExpression:      expr.Projection(),
		})
		if err != nil {
			log.Printf("Couldn't scan for movies released between")
		} else {
			err = attributevalue.UnmarshalListOfMaps(response.Items, &shipments)
			if err != nil {
				log.Printf("Couldn't unmarshal query response. Here's why: %v\n", err)
			}
		}
	}
	return shipments
}

func (basics TableBasics) ScanCourier() []Courier {
	var err error
	var couriers []Courier
	filtEx := expression.Name("available").Equal(expression.Value(true))
	projEx := expression.NamesList(
		expression.Name("ID"), expression.Name("x"), expression.Name("y"))
	expr, err := expression.NewBuilder().WithFilter(filtEx).WithProjection(projEx).Build()
	if err != nil {
		log.Printf("Couldn't build expressions for scan. Here's why: %v\n", err)
	} else {
		var response *dynamodb.ScanOutput
		response, err = basics.DynamoDbClient.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName:                 aws.String(basics.TableName),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ProjectionExpression:      expr.Projection(),
		})
		if err != nil {
			log.Printf("Couldn't scan for movies released between")
		} else {
			err = attributevalue.UnmarshalListOfMaps(response.Items, &couriers)
			if err != nil {
				log.Printf("Couldn't unmarshal query response. Here's why: %v\n", err)
			}
		}
	}
	return couriers
}

func assign(w http.ResponseWriter, req *http.Request) {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	courierTable := TableBasics{TableName: "courier_management",
		DynamoDbClient: dynamodb.NewFromConfig(sdkConfig)}

	couriers := courierTable.ScanCourier()
	fmt.Print(couriers)

	shipmentTable := TableBasics{TableName: "shipment_management",
		DynamoDbClient: dynamodb.NewFromConfig(sdkConfig)}
	shipments := shipmentTable.ScanShipments("2022/12/12")
	fmt.Print(shipments)

}

func calculate(w http.ResponseWriter, req *http.Request) {

}

func main() {
	http.HandleFunc("/assign", assign)
	http.HandleFunc("/calculate", calculate)

	err := http.ListenAndServe("localhost:8888", nil)

	if err != nil {
		log.Fatal(err)
	}
}
