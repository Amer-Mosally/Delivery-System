HOW TO RUN:

1. Initiate the infrastructure by running 'terraform init'
2. Initilize the database by running the postman collections. Make sure to change the apigetway varible to the provisioned APIGetway's url.
3. Access the ec2 instance using SSH. Make sure you create a key with the name "cx-project.pem" attached to your account. download this key and place it in the folder. Then run the following commands there:
"export AWS_ACCESS_KEY_ID=*****************
export AWS_SECRET_ACCESS_KEY=********************
export AWS_REGION=us-east-1
my_ip=$(curl http://checkip.amazonaws.com)
export my_ip"
4. Trigger the assign function by hitting the end point: ec2_public_ip/assign with this body:
{
"date": "year/month/day"
}
5. Run the workload provided, or as instructed in the project report
