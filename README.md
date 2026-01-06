## Note: 
‼️ Please run the "inshorts-news-data-syncer" to create Elasticsearch index along with settings and mappings before running "inshorts-news-srv"

## Steps run this app
1. Install ElasticSearch in your machine and start it before running this app
2. Set up a go runtime as this app is written in Golang
3. Once Golang setup is done you can start this app by running this command:-
- go run cmd/main.go 
4. Import the postman collection attached with the mail/project

## Technology Choices

This project is built using Golang, Elasticsearch, Fiber, and OpenRouter (LLM) to deliver a fast, intelligent, and scalable news search platform.

* Golang is used for its high performance, efficient concurrency model, and reliability in building production-grade backend services.

* Elasticsearch powers full-text search, relevance scoring, geo-spatial queries, and real-time analytics, which are critical for news discovery.

* Fiber provides a lightweight, high-performance HTTP framework with low latency and minimal overhead.

* OpenRouter (LLM) is integrated to intelligently analyze user queries, detect intent, extract entities, and generate concise summaries, enabling smarter and more relevant search experiences.

Together, these technologies create a scalable, search-optimized system with AI-driven understanding

‼️ The OpenRouter LLM API may introduce 2–3 seconds of latency for certain requests. To avoid impacting core search performance, LLM usage is designed to be fail-soft and non-blocking, and is primarily used for query understanding and result enrichment, not for primary data retrieval.

‼️ To test trending API please take the lat and lon from the stream simulated logger and then hit the APIs as for now I have tried adding user events randomly