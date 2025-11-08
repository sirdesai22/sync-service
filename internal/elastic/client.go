package elastic

import (
	"log"
	"os"

	es "github.com/elastic/go-elasticsearch/v8"
)

func Connect() *es.Client {
	cfg := es.Config{
		Addresses: []string{os.Getenv("ELASTIC_URL")},
	}
	client, err := es.NewClient(cfg)
	if err != nil {
		log.Fatalf("❌ failed to connect to Elasticsearch: %v", err)
	}
	log.Println("✅ Connected to Elasticsearch")
	return client
}
