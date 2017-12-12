package es

import (
	"sync"
	"time"
	"fmt"
	"strings"
	"log"
	"encoding/json"
)

const (
	DefaultIndexPatternStr	= "index-pattern"
	DefaultSyncIndexPatternTime = 120
)

type Index struct {
	Index	string		`json:"index"`
	Status	string		`json:"status"`
}

type IndexPatterns struct {
	patterns		[]string
	mutex			*sync.RWMutex
}

func NewIndexPatterns() *IndexPatterns {
	return &IndexPatterns{
		patterns: []string{},
		mutex: &sync.RWMutex{},
	}
}

func (p *IndexPatterns) GetIndexPatterns() []string {
	patterns := []string{}
	p.mutex.RLock()
	for _, pattern := range p.patterns {
		patterns = append(patterns, pattern)
	}
	p.mutex.RUnlock()
	return patterns
}

func (p *IndexPatterns) SetIndexPatterns(patterns []string) {
	p.mutex.Lock()
	p.patterns = patterns
	p.mutex.Unlock()
}


type KibanaSimilar struct {
	Client			*Client
	kibanaIndex		string
	Indices			[]Index
	Patterns		*IndexPatterns
}

func NewKibanaSimilar(host, kibanaIndex string) (*KibanaSimilar, error) {
	ks := &KibanaSimilar{
		Client: NewClient(host),
		kibanaIndex: kibanaIndex,
		Patterns: NewIndexPatterns(),
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

	log.Println("Load Indices: ", ks.Indices)
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

func (ks *KibanaSimilar) SyncIndexPatterns() error {
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

	ks.Patterns.SetIndexPatterns(searchResult.GetIndexPatterns())
	log.Println("Load Index Patterns: ", ks.Patterns.GetIndexPatterns())
	return nil
}

func (ks *KibanaSimilar) LoadKibanaIndexPatterns() error {
	if !ks.IsKibanaIndexExists() {
		if err := ks.CreateKibanaIndex(); err != nil {
			return err
		}
	}

	if err := ks.SyncIndexPatterns(); err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(DefaultSyncIndexPatternTime * time.Second)
		for _ = range ticker.C {
			if err := ks.SyncIndexPatterns(); err != nil {
				log.Println("Sync Index Patterns failed! Error: ", err)
			}
		}
	}()
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
	for _, pattern := range ks.Patterns.GetIndexPatterns() {
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
