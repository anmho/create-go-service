#!/usr/bin/env bash

# DynamoDB Local endpoint
DYNAMO_ENDPOINT="http://localhost:8000"

# Table name
TABLE_NAME="NoteTable"

# Create the table
aws dynamodb create-table \
    --table-name "$TABLE_NAME" \
    --attribute-definitions \
        AttributeName=UserId,AttributeType=S \
        AttributeName=Timestamp,AttributeType=S \
    --key-schema \
        AttributeName=UserId,KeyType=HASH \
        AttributeName=Timestamp,KeyType=RANGE \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --endpoint-url "$DYNAMO_ENDPOINT"

echo "Table '$TABLE_NAME' created on DynamoDB Local at $DYNAMO_ENDPOINT"
