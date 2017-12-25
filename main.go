package main

import (
	"strings"
	"strconv"
	"time"
	"log"
	"net/http"
	"encoding/json"
	"os"

	"github.com/aliate/autopattern/es.client"
)

var kibanaSimilar *es.KibanaSimilar

type Config struct {
	Port			string
	ESHosts			[]string
	ESIndexKeepDays	int
	KibanaIndex		string
}

const (
	DefaultPort = "12345"
	DefaultIndexKeepDays = 30
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
	esIndexKeepDays := os.Getenv("ES_INDEX_KEEP_DAYS")
	if len(esIndexKeepDays) > 0 {
		keepDays, err := strconv.Atoi(esIndexKeepDays)
		if err != nil {
			log.Printf("ES_INDEX_KEEP_DAYS parse failed! Error: %s\n", err)
			log.Printf("Use default 30 days...\n")
			keepDays = DefaultIndexKeepDays
		}
		c.ESIndexKeepDays = keepDays
	} else {
		c.ESIndexKeepDays = DefaultIndexKeepDays
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

const (
	banner = `
   ___          __                    __    __               
  /   | __ ____/ /____  ___  __/\  __/ /___/ /____  ___  __  
 / _| |/ // /_  __/ _ \/ _ \/ _ /_/_  __/_  __/ __\/ __\/ _ \
/_/ |_|\____//__/ \___/ ___/\____/ /__/  /__/ \___/_/  /_//_/  %s
                     /_/                                     
	`
	version = "1.1.0"
)

func showBanner() {
	log.Printf(banner, version)
}

func main() {
	showBanner()

	var config Config
	config.Load()
	if len(config.ESHosts) == 0 {
		log.Fatal("ES_HOSTS must been set in env!")
	}

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

	// Start Curator for auto delete indices
	curator := es.NewCurator(config.ESHosts[0], config.ESIndexKeepDays)
	if err := curator.Start(); err != nil {
		log.Panicf("Start Curator failed! Error: %s\n", err)
	}

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":" + config.Port, nil))
}
