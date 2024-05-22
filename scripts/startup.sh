#!/bin/sh
docker-compose up -d elasticsearch kafka init-kafka kafka-connect

# Wait till Elasticsearch is up
until curl -sS 'http://localhost:9200/_cat/health?h=status' | grep -q 'green\|yellow'; do
  sleep 1
done

# Create Elasticsearch topic
curl -X PUT 'http://localhost:9200/users' -H 'Content-Type: application/json' -d'
{
  "mappings": {
	"properties": {
	  "id": { "type": "text" },
	  "username": { "type": "text" },
	  "email": { "type": "text" }
	}
  }
}'


# blocks until kafka is reachable
kafka-topics --bootstrap-server localhost:9092 --list

curl -X POST http://localhost:8083/connectors -H 'Content-Type: application/json' -d \
'{
  "name": "elasticsearch-sink",
  "config": {
    "connector.class": "io.confluent.connect.elasticsearch.ElasticsearchSinkConnector",
    "tasks.max": "1",
    "topics": "users",
	"key.ignore": "false",
    "schema.ignore": "true",
    "connection.url": "http://elastic:9200",
    "type.name": "_doc",
    "name": "elasticsearch-sink",
    "value.converter": "org.apache.kafka.connect.json.JsonConverter",
    "value.converter.schemas.enable": "false",
	"behaviour.on.null.vallues": "DELETE"
  }
}'

docker-compose up
