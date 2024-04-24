# Review Chatbot

## Table of Contents
- [Requirements](#requirements)
- [Getting Started](#getting-started)

## Requirements

Golang & docker-compose

## Getting Started

```
docker-compose up
go run cmd/chat-server/main.go
```
The chatbot implementation is using openAI, so please replace the OPENAI_TOKEN value in the .env file

* UI will be on localhost:80
* Grafana is on localhost:3000 (admin/grafana)
* Uptrace is on localhost:14318 (default credentials)

* To trigger the review:
```
curl -X POST localhost:8181/api/v1/review --data '{"userId": 2, "productId": 3}'
```
