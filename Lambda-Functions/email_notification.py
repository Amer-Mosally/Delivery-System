import json
import boto3

def lambda_handler(event, context):
    data = json.loads(event['body'])
    send = data['send']

    method = event['httpMethod']

    # this will create dynamodb resource object and 'dynamodb' is resource name
    dynamodb = boto3.resource('dynamodb')
    # this will search for dynamoDB table 
    table = dynamodb.Table("result")
    
    if method == "GET":
        all = table.scan()
        all = str(all['Items'])  #retrun list of items
        
        client = boto3.client("ses")
        subject = "Result from the database"

        message = {"Subject": {"Data" : subject},
                    "Body": {"Html": {"Data": all}}}

        response = client.send_email(Source = "musalli.amer@gmail.com",
                Destination = {"ToAddresses": ["musalli.amer@gmail.com"]}, Message = message)
    
    return {
        'statusCode': 200,
        'body': json.dumps("The email has been send!")
    }