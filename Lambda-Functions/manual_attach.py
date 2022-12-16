import json
import boto3

def lambda_handler(event, context):
    data = json.loads(event['body'])
    PackageID = data['PackageID']   
    CourierID = data['CourierID']   
    x = data['x']
    y = data['y']

    method = event['httpMethod']

    # this will create dynamodb resource object and 'dynamodb' is resource name
    dynamodb = boto3.resource('dynamodb')
    # this will search for dynamoDB table 
    table = dynamodb.Table("result")

    if method == "GET":
        all = table.scan()
        all = all['Items']  #retrun list of items
        JsonValue = str(all)

    if method == "POST":
        try:
            response = table.get_item(Key={'PackageID': PackageID})
            response = response['Item']['PackageID']
        except Exception as error:
            table.put_item(Item={'PackageID': PackageID, 'CourierID': CourierID,'x':x,'y':y})
            JsonValue = "The ID has been added to the database!"
        if response == PackageID:
                JsonValue = "The Package ID already in the database!"

                
    if method == "DELETE":
        #test
        JsonValue = "The result table has been deleted"
    
    return {
        'statusCode': 200,
        'body': json.dumps(JsonValue)
    }
  