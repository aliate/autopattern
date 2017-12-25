package es

import (
	"time"
	"encoding/json"
	"log"
)

type Curator struct {
	Client			*Client
	IndexKeepDays	int
	ClearTime		int
	stop			chan struct{}
}

const (
	DefaultClearTime = 0
	DefaultTimeFormat = "2016-01-02"
)

func NewCurator(esHost string, keepDays int) *Curator {
	return &Curator{
		Client: NewClient(esHost),
		IndexKeepDays: keepDays,
		stop: make(chan struct{}),
	}
}

func (c *Curator) deleteIndex(index string) error {
	_, err := c.Client.Delete(index)
	return err
}

func (c *Curator) clearPassedIndices() {
	body, err := c.Client.Get("/_cat/indices")
	if err != nil {
		log.Printf("Curator get indices failed! Error: %s\n", err)
		return
	}

	indices := []Index{}
	err = json.Unmarshal(body, &indices)
	if err != nil {
		log.Printf("Curator Parse indices body failed! Error: %s\n", err)
		return
	}

	getPassedDays := func(index string) int {
		indexTime := index[len(index) - len(DefaultTimeFormat):]
		then, err := time.Parse(DefaultTimeFormat, indexTime)
		if err != nil {
			log.Printf("Parse %s failed! Error: %s\n", indexTime, err)
			return 0
		}
		return int(time.Since(then).Hours()/24)
	}

	for _, index := range indices {
		if getPassedDays(index.Index) > c.IndexKeepDays {
			err = c.deleteIndex(index.Index)
			if err != nil {
				log.Printf("Delete index %s failed! Error: %s\n", index.Index, err)
			}
		}
	}
}

func (c *Curator) Start() error {
	go func() {
		ticker := time.NewTicker(time.Hour * 1)
		for {
			select {
			case t := <-ticker.C:
				if t.Hour() == DefaultClearTime {
					c.clearPassedIndices()
				}
			case <-c.stop:
				return
			}
		}
	}()
	return nil
}

func (c *Curator) Stop() error {
	c.stop <- struct{}{}
	return nil
}
