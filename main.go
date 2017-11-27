package main

import (
	"time"
	"log"
	"net/http"
	"encoding/json"
	"os"

	"github.com/aliate/autopattern/es"
)

var esClient *es.ESClient

type Config struct {
	Port			string
	ESHosts		[]string
}

const (
	defaultPort = "12345"
)

func (c *Config) Load() {
	akPort := os.Getenv("AK_PORT")
	if len(akPort) > 0 {
		c.Port = akPort
	} else {
		c.Port = defaultPort
	}
	esHosts := os.Getenv("ES_HOSTS")
	if len(esHosts) > 0 {
		err := json.Unmarshal([]byte(esHosts), &c.ESHosts)
		if err != nil {
			log.Printf("Unmarshal ES_HOSTS Error: %s\n", err)
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "POST" {
		indexPattern := r.FormValue("index-pattern")
		if len(indexPattern) > 0 {
			processIndexPattern(indexPattern)
		}
	}
}

func processIndexPattern(indexPattern string) {
	err := esClient.ProcessIndexPattern(indexPattern)
	if err != nil {
		log.Printf("Process index-pattern: %s Error: %s\n", indexPattern, err)
	}
}

func main() {
	var config Config
	config.Load()
	if len(config.ESHosts) == 0 {
		log.Fatal("ES_HOSTS must been set in env!")
	}

	for {
		esClient = es.NewESClient(config.ESHosts[0])
		if esClient != nil {
			break
		}
		log.Println("Create es client Failed! esHost: ", config.ESHosts[0])
		time.Sleep(30 * time.Second)
	}

	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":" + config.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
