package es

type Source struct {
	ID		string		`json:"_id"`
	Type	string		`json:"_type"`
}

type Hits struct {
	Total	int			`json:"total"`
	Hits	[]Source	`json:"hits"`
}

type SearchResult struct {
	Hits	Hits		`json:"hits"`
}


func (s *SearchResult) GetIndexPatterns() []string {
	indexPatterns := []string{}
	for _, source := range s.Hits.Hits {
		if source.Type == DefaultIndexPatternStr {
			indexPatterns = append(indexPatterns, source.ID)
		}
	}
	return indexPatterns
}
