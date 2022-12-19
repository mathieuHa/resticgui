package Models

type Matches struct {
	Snapshot string  `json:"snapshot"`
	Hits     int     `json:"hits"`
	Files    []Files `json:"matches"`
}
