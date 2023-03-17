#!/bin/bash

MONGO_TEST_PORT=27017
MONGO_IMAGE="mongo:6"
CONTAINER_NAME="mongo"
MONGO_URI=mongodb://localhost:$MONGO_TEST_PORT
APP_SRC=$(cat .env | grep "APP_SRC" | cut -d "=" -f2)
DB_NAME="poison"

export MONGO_URI=$MONGO_URI
export DB_NAME=$DB_NAME
# run mongo
CONTAINER_ID=$(docker run --rm -d -p $MONGO_TEST_PORT:27017 --name=$CONTAINER_NAME -e MONGODB_DATABASE=$DB_NAME $MONGO_IMAGE)
echo $CONTAINER_ID
# run migrations
#docker run -v $APP_SRC/migrations:/migrations --network host --rm migrate/migrate -path=/migrations/ -database $MONGO_URI/$DB_NAME up
