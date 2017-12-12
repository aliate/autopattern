package main

import (
	"strings"
	"time"
	"log"
	"net/http"
	"encoding/json"
	"os"

	"github.com/aliate/autopattern/es.client"
)

var kibanaSimilar *es.KibanaSimilar

type Config struct {
	Port		string
	ESHosts		[]string
	KibanaIndex	string
}

const (
	DefaultPort = "12345"
	DefaultKibanaIndex = ".kibana"
)

func processHosts(hosts string) string {
	if strings.Contains(hosts, "'") {
		return strings.Replace(hosts, "'", "\"", -1)
	}
	return hosts
}

func (c *Config) Load() {
	akPort := os.Getenv("AK_PORT")
	if len(akPort) > 0 {
		c.Port = akPort
	} else {
		c.Port = DefaultPort
	}
	esHosts := os.Getenv("ES_HOSTS")
	if len(esHosts) > 0 {
		err := json.Unmarshal([]byte(processHosts(esHosts)), &c.ESHosts)
		if err != nil {
			log.Printf("Unmarshal ES_HOSTS Error: %s\n", err)
		}
	}
	kibanaIndex := os.Getenv("KIBANA_INDEX")
	if len(kibanaIndex) > 0 {
		c.KibanaIndex = kibanaIndex
	} else {
		c.KibanaIndex = DefaultKibanaIndex
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
	err := kibanaSimilar.ProcessIndexPattern(indexPattern)
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

	log.Println("Welcome enter autopattern!")
	log.Printf("Config: %#v\n", config)

	for {
		var err error
		kibanaSimilar, err = es.NewKibanaSimilar(config.ESHosts[0], config.KibanaIndex)
		if err == nil {
			break
		}
		log.Printf("Connect es: %s client Failed! Error: %s\n", config.ESHosts[0], err)
		time.Sleep(30 * time.Second)
	}
	kibanaSimilar.InitKibanaIndexPatterns()

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":" + config.Port, nil))
}
