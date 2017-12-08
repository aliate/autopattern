package es

import (
	"fmt"
	"strings"
	"encoding/json"
)

const (
	DefaultIndexPatternStr	= "index-pattern"
)

type Index struct {
	Index	string		`json:"index"`
	Status	string		`json:"status"`
}

type KibanaSimilar struct {
	Client		*Client
	kibanaIndex	string
	Indices		[]Index
	Patterns	[]string
}

func NewKibanaSimilar(host, kibanaIndex string) (*KibanaSimilar, error) {
	ks := &KibanaSimilar{
		Client: &Client{
			Host: host,
		},
		kibanaIndex: kibanaIndex,
	}

	if err := ks.LoadIndices(); err != nil {
		return nil, err
	}

	if err := ks.LoadKibanaIndexPatterns(); err != nil {
		return nil, err
	}

	return ks, nil
}

func (ks *KibanaSimilar) LoadIndices() error {
	body, err := ks.Client.Get("/_cat/indices")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &ks.Indices); err != nil {
		return err
	}

	return nil
}

func (ks *KibanaSimilar) IsKibanaIndexExists() bool {
	for _, index := range ks.Indices {
		if index.Index == ks.kibanaIndex {
			return true
		}
	}
	return false
}

func (ks *KibanaSimilar) CreateKibanaIndex() error {
	_, err := ks.Client.Put("/" + ks.kibanaIndex)
	if err != nil {
		return err
	}
	return nil
}

func (ks *KibanaSimilar) GetKibanaIndexPatterns() ([]string, error) {
	body, err := ks.Client.Get(
		fmt.Sprintf("/%s/%s/_search", ks.kibanaIndex, DefaultIndexPatternStr))
	if err != nil {
		return nil, err
	}

	var searchResult SearchResult
	err = json.Unmarshal(body, &searchResult)
	if err != nil {
		return nil, err
	}

	var patterns []string
	if searchResult.Hits.Total > 0 {
		for _, source := range searchResult.Hits.Hits {
			if source.Type == DefaultIndexPatternStr {
				patterns = append(patterns, source.ID)
			}
		}
	}
	return patterns, nil
}

func (ks *KibanaSimilar) LoadKibanaIndexPatterns() error {
	if !ks.IsKibanaIndexExists() {
		if err := ks.CreateKibanaIndex(); err != nil {
			return err
		}
	}

	body, err := ks.Client.Get(
		fmt.Sprintf("/%s/%s/_search", ks.kibanaIndex, DefaultIndexPatternStr))
	if err != nil {
		return err
	}

	var searchResult SearchResult
	err = json.Unmarshal(body, &searchResult)
	if err != nil {
		return err
	}

	var patterns []string
	if searchResult.Hits.Total > 0 {
		for _, source := range searchResult.Hits.Hits {
			if source.Type == DefaultIndexPatternStr {
				patterns = append(patterns, source.ID)
			}
		}
	}
	ks.Patterns = patterns

	return nil
}

func (ks *KibanaSimilar) CreateKibanaIndexPattern(indexPattern string) error {
	var data = struct {
		Title	string	`json:"title"`
	}{
		Title: indexPattern,
	}
	dataBytes, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	dataReader := strings.NewReader(string(dataBytes))
	_, err = ks.Client.Post(
		fmt.Sprintf("/%s/%s/%s", ks.kibanaIndex, DefaultIndexPatternStr, indexPattern),
		dataReader)
	if err != nil {
		return err
	}
	return nil
}

func (ks *KibanaSimilar) IsKibanaIndexPatternExists(indexPattern string) bool {
	for _, pattern := range ks.Patterns {
		if pattern == indexPattern {
			return true
		}
	}
	return false
}

func (ks *KibanaSimilar) ProcessIndexPattern(indexPattern string) error {
	if !ks.IsKibanaIndexPatternExists(indexPattern) {
		if err := ks.CreateKibanaIndexPattern(indexPattern); err != nil {
			return err
		}
	}
	return nil
}
