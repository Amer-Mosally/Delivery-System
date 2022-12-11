import json
import boto3

def lambda_handler(event, context):
    data = json.loads(event['body'])
    x = data['x']   #json.loads(event['x'])
    y = data['y']    #json.loads(event['y'])
    
    ID =str(x)+"div"+str(y)  #Stored in DB as "1add2"

    # this will create dynamodb resource object and dynamodb is resource name
    dynamodb = boto3.resource('dynamodb')
    # this will search for dynamoDB table 
    table = dynamodb.Table("HW3_table")
    
    
    try:
        response = table.get_item(Key={'ID': ID})
        res = int(response['Item']['result'])
    except Exception as error:
        print("No item! Adding new Item to database")
        res = int(x) / int(y)
        table.put_item(Item={'ID': ID,'result':res})

    JsonValue = {
        "x": x,
        "y": y,
        "op": "div",
        "result": res
    }
    return {
        'statusCode': 200,
        'body': json.dumps(JsonValue)
    }
  