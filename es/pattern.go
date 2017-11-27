package es

type Pattern struct {
	Title	string		`json:"title"`
}

type Index struct {
	Type	string		`json:"_type"`
	Source Pattern	 `json:"_source"`
}

type Hits struct {
	Total	int			  `json:"total"`
	Hits	[]Index		`json:"hits"`
}

type SearchResult struct {
	Hits	Hits			`json:"hits"`
}