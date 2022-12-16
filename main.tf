provider "aws" {
  region     = "us-east-1"
  access_key = ""
  secret_key = ""
}
data "archive_file" "lambda-functions" {
  type        = "zip"
  source_dir  = "./Lambda-Functions"
  output_path = "./Lambda-Functions.zip"
}

### SES ###

resource "aws_ses_email_identity" "email" {
  email = "musalli.amer@gmail.com"
}
### Policy ###
data "aws_iam_policy_document" "assume_role" {
  statement {
    sid    = ""
    effect = "Allow"
    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }
    actions = ["sts:AssumeRole",
    ]
  }
}
data "aws_iam_policy_document" "DynamoDB" {
  statement {
    actions = [
      "dynamodb:PutItem",
      "dynamodb:GetItem",
      "dynamodb:DeleteItem",
      "dynamodb:Scan",
      "ses:SendEmail"
    ]
    effect    = "Allow"
    resources = ["*"]
  }
}
resource "aws_iam_role_policy_attachment" "assume_role" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.iam_for_lambda.name
}
resource "aws_iam_role" "iam_for_lambda" {
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
  name               = "IAMPolicyforDynamoDB"
}
resource "aws_iam_policy_attachment" "DynamoDB" {
  name       = "IAMPolicyforDynamoDB"
  policy_arn = aws_iam_policy.DynamoDB.arn
  roles      = [aws_iam_role.iam_for_lambda.name]
}
resource "aws_iam_policy" "DynamoDB" {
  name   = "IAMPolicyforDynamoDB"
  policy = data.aws_iam_policy_document.DynamoDB.json
}

### lambda ###

resource "aws_lambda_function" "lambda-courier_management" {
  function_name    = "courier_management"
  filename         = data.archive_file.lambda-functions.output_path
  source_code_hash = data.archive_file.lambda-functions.output_base64sha256
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "courier_management.lambda_handler"
  runtime          = "python3.9"
}
resource "aws_lambda_function" "lambda-shipment_management" {
  function_name    = "shipment_management"
  filename         = data.archive_file.lambda-functions.output_path
  source_code_hash = data.archive_file.lambda-functions.output_base64sha256
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "shipment_management.lambda_handler"
  runtime          = "python3.9"
}
resource "aws_lambda_function" "lambda-manual_attach" {
  function_name    = "manual_attach"
  filename         = data.archive_file.lambda-functions.output_path
  source_code_hash = data.archive_file.lambda-functions.output_base64sha256
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "manual_attach.lambda_handler"
  runtime          = "python3.9"
}
resource "aws_lambda_function" "lambda-email_notification" {
  function_name    = "email_notification"
  filename         = data.archive_file.lambda-functions.output_path
  source_code_hash = data.archive_file.lambda-functions.output_base64sha256
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "email_notification.lambda_handler"
  runtime          = "python3.9"
}

### Gateway ###

resource "aws_apigatewayv2_api" "lambda" {
  name          = "serverless_lambda_gw"
  protocol_type = "HTTP"
}
resource "aws_apigatewayv2_stage" "lambda" {
  api_id      = aws_apigatewayv2_api.lambda.id
  name        = "serverless_lambda_stage"
  auto_deploy = true
}

### Integrattion ###

# courier_management
resource "aws_apigatewayv2_integration" "courier_management" {
  api_id = aws_apigatewayv2_api.lambda.id

  integration_uri    = aws_lambda_function.lambda-courier_management.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST" # REST API communication between API Gateway and Lambda
}
resource "aws_apigatewayv2_route" "courier_management" {
  api_id = aws_apigatewayv2_api.lambda.id

  route_key = "ANY /courier_management" #accept ANY method (get, post...)
  target    = "integrations/${aws_apigatewayv2_integration.courier_management.id}"
}

# shipment_management
resource "aws_apigatewayv2_integration" "shipment_management" {
  api_id = aws_apigatewayv2_api.lambda.id

  integration_uri    = aws_lambda_function.lambda-shipment_management.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST" # REST API communication between API Gateway and Lambda
}
resource "aws_apigatewayv2_route" "shipment_management" {
  api_id = aws_apigatewayv2_api.lambda.id

  route_key = "ANY /shipment_management"
  target    = "integrations/${aws_apigatewayv2_integration.shipment_management.id}"
}

# manual_attach
resource "aws_apigatewayv2_integration" "manual_attach" {
  api_id = aws_apigatewayv2_api.lambda.id

  integration_uri    = aws_lambda_function.lambda-manual_attach.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST" # REST API communication between API Gateway and Lambda
}
resource "aws_apigatewayv2_route" "manual_attach" {
  api_id = aws_apigatewayv2_api.lambda.id

  route_key = "ANY /manual_attach"
  target    = "integrations/${aws_apigatewayv2_integration.manual_attach.id}"
}

#SES
resource "aws_apigatewayv2_integration" "email_notification" {
  api_id = aws_apigatewayv2_api.lambda.id

  integration_uri    = aws_lambda_function.lambda-email_notification.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST" # REST API communication between API Gateway and Lambda
}
resource "aws_apigatewayv2_route" "email_notification" {
  api_id = aws_apigatewayv2_api.lambda.id

  route_key = "ANY /email_notification" #accept ANY method (get, post...)
  target    = "integrations/${aws_apigatewayv2_integration.email_notification.id}"
}

### Lambda Permission ###

resource "aws_lambda_permission" "courier_management" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda-courier_management.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.lambda.execution_arn}/*/*"
}
resource "aws_lambda_permission" "api_gw_shipment_management" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda-shipment_management.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.lambda.execution_arn}/*/*"
}
resource "aws_lambda_permission" "api_gw_manual_attach" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda-manual_attach.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.lambda.execution_arn}/*/*"
}
resource "aws_lambda_permission" "api_gw_email_notification" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda-email_notification.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.lambda.execution_arn}/*/*"
}

### Creating table ###

resource "aws_dynamodb_table" "courier_management" {
  name           = "courier_management"
  billing_mode   = "PROVISIONED"
  read_capacity  = "5"
  write_capacity = "5"
  hash_key       = "ID"
  attribute {
    name = "ID"
    type = "N" #int
  }
}
resource "aws_dynamodb_table" "shipment_management" {
  name           = "shipment_management"
  billing_mode   = "PROVISIONED"
  read_capacity  = "5"
  write_capacity = "5"
  hash_key       = "ID"
  attribute {
    name = "ID"
    type = "N" #int
  }
}
resource "aws_dynamodb_table" "result" {
  name           = "result"
  billing_mode   = "PROVISIONED"
  read_capacity  = "5"
  write_capacity = "5"
  hash_key       = "PackageID"
  attribute {
    name = "PackageID"
    type = "N" #int
  }
}
