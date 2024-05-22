#!/bin/sh
docker-compose up -d elasticsearch

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

docker-compose up
