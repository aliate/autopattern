package es

import (
	"time"
	"strings"
	"log"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

type ESClient struct {
	esHost		string
	indices		[]string
	patterns   []string
}

func NewESClient(esHost string) *ESClient {
	client := &ESClient{
		esHost: "http://" + esHost,
		indices: []string{},
		patterns: []string{},
	}
	if err := client.loadIndices(); err != nil {
		log.Printf("Load indices from es: %s Error: %s\n", esHost, err)
		return nil
	}
	if err := client.loadKibanaPatterns(); err != nil {
		log.Printf("Load kibana patterns from es: %s Error: %s\n", esHost, err)
		return nil
	}
	return client
}

func (e *ESClient) loadIndices() error {
	resp, err := http.Get(e.esHost + "/_mapping")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	mapping := make(map[string]interface{})
	err = json.Unmarshal(result, &mapping)
	if err != nil {
		return err
	}

	var indices []string
	for k, _ := range mapping {
		indices = append(indices, k)
	}
	e.indices = indices
	return nil
}

func (e *ESClient) loadKibanaPatterns() error {
	var hasKibanaIndices bool
	for _, index := range e.indices {
		if index == ".kibana" {
			hasKibanaIndices = true
		}
	}

	if !hasKibanaIndices {
		if err := e.createKibanaIndex(); err != nil {
			return err
		}
	}

	if err := e.getKibanaPatterns(); err != nil {
		return err
	}

	go func() {
		tick := time.NewTicker(60 * time.Second)
		for _ = range tick.C {
			e.getKibanaPatterns()
		}
	}()

	return nil
}

func (e *ESClient) getKibanaPatterns() error {
	resp, err := http.Get(e.esHost + "/.kibana/index-pattern/_search")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var searchResult SearchResult
	err = json.Unmarshal(result, &searchResult)
	if err != nil {
		return err
	}

	var patterns []string
	if searchResult.Hits.Total > 0 {
		for _, i := range searchResult.Hits.Hits {
			patterns = append(patterns, i.Source.Title)
		}
	}
	e.patterns = patterns
	log.Printf("Get kibana index Patterns: %v \n", e.patterns)
	return nil
}

func (e *ESClient) createKibanaIndex() error {
	client := &http.Client{}
	reqest, err := http.NewRequest("PUT", e.esHost + "/.kibana", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(reqest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Println("Created .kibana index in es")
	return nil
}

func (e *ESClient) createKibaneIndexPattern(indexPattern string) error {
	var body = struct {
		Title 	string 		`json:"title"`
	}{
		Title: indexPattern,
	}
	bs, err := json.Marshal(&body)
	if err != nil {
		return err
	}
	b := strings.NewReader(string(bs))
	resp, err := http.Post(e.esHost + "/.kibana/index-pattern/" + indexPattern, "application/json", b)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (e *ESClient) ProcessIndexPattern(pattern string) error {
	var exists bool
	if len(e.patterns) > 0 {
		for _, p :=range e.patterns {
			if p == pattern {
				exists = true
			}
		}
	}
	if !exists {
		if err := e.createKibaneIndexPattern(pattern); err != nil {
			return err
		}
		e.patterns = append(e.patterns, pattern)
		log.Printf("Create new .kibana index-pattern: %s\n", pattern)
	}
	return nil
}