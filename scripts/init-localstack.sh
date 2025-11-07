#!/bin/bash

# Initialize LocalStack S3 bucket for development
# This script creates the required S3 bucket and configures CORS

set -e

ENDPOINT_URL="http://localhost:4566"
BUCKET_NAME="cashlens-uploads-dev"
REGION="ap-south-1"

echo "Waiting for LocalStack to be ready..."
until curl -s "${ENDPOINT_URL}/_localstack/health" | grep -q '"s3":"available"'; do
  sleep 1
done
echo "LocalStack is ready!"

# Check if bucket exists
if aws --endpoint-url=$ENDPOINT_URL s3 ls s3://$BUCKET_NAME 2>/dev/null; then
  echo "Bucket $BUCKET_NAME already exists"
else
  echo "Creating bucket $BUCKET_NAME..."
  aws --endpoint-url=$ENDPOINT_URL s3 mb s3://$BUCKET_NAME --region $REGION
  echo "Bucket created successfully"
fi

# Configure CORS
echo "Configuring CORS for bucket..."
cat > /tmp/cors-config.json <<EOF
{
  "CORSRules": [
    {
      "AllowedOrigins": ["http://localhost:3000"],
      "AllowedMethods": ["GET", "PUT", "POST", "DELETE", "HEAD"],
      "AllowedHeaders": ["*"],
      "ExposeHeaders": ["ETag"],
      "MaxAgeSeconds": 3000
    }
  ]
}
EOF

aws --endpoint-url=$ENDPOINT_URL s3api put-bucket-cors \
  --bucket $BUCKET_NAME \
  --cors-configuration file:///tmp/cors-config.json

echo "CORS configured successfully"
echo "LocalStack initialization complete!"
