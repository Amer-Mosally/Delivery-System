package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"

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
	ID       int    `dynamodbav:"ID"`
	X        string `dynamodbav:"x"`
	Y        string `dynamodbav:"y"`
	distance float64
}

type Shipment struct {
	ID       int    `dynamodbav:"ID"`
	X        string `dynamodbav:"x"`
	Y        string `dynamodbav:"y"`
	distance float64
}

type AssignedShipment struct {
	shipmentID int
	courierID  int
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

func assignDriver(couriers []Courier, shipments []Shipment) []AssignedShipment {
	for i := 0; i < len(couriers); i++ {
		a, _ := strconv.ParseFloat(couriers[i].X, 64)
		b, _ := strconv.ParseFloat(couriers[i].Y, 64)
		couriers[i].distance = math.Sqrt(math.Pow(a, 2) + math.Pow(b, 2))
	}

	for i := 0; i < len(shipments); i++ {
		a, _ := strconv.ParseFloat(shipments[i].X, 64)
		b, _ := strconv.ParseFloat(shipments[i].Y, 64)
		shipments[i].distance = math.Sqrt(math.Pow(a, 2) + math.Pow(b, 2))
	}

	sort.Slice(couriers, func(i, j int) bool {
		return couriers[i].distance > couriers[j].distance
	})

	sort.Slice(shipments, func(i, j int) bool {
		return shipments[i].distance > shipments[j].distance
	})

	var assignedShipments []AssignedShipment

	for i := 0; i < len(shipments); i++ {
		assignedShipments = append(assignedShipments, AssignedShipment{
			shipmentID: shipments[i].ID,
			courierID:  couriers[i].ID,
		})
	}

	return assignedShipments
}

func assign(w http.ResponseWriter, req *http.Request) {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	reqBody, _ := ioutil.ReadAll(req.Body) //extracting the date from the request
	var jsonBody map[string]string
	json.Unmarshal(reqBody, &jsonBody)
	date := jsonBody["date"]

	courierTable := TableBasics{TableName: "courier_management",
		DynamoDbClient: dynamodb.NewFromConfig(sdkConfig)}

	couriers := courierTable.ScanCourier() //retrieving all couriers that are available

	shipmentTable := TableBasics{TableName: "shipment_management",
		DynamoDbClient: dynamodb.NewFromConfig(sdkConfig)}

	shipments := shipmentTable.ScanShipments(date) //retrieving all shipments on this date

	assignedShipments := assignDriver(couriers, shipments)
	fmt.Println(assignedShipments)
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
