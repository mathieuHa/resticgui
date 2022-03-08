package Models

import "time"

type Snapshot struct {
	Time     time.Time `json:"time"`
	Tree     string    `json:"tree"`
	Hostname string    `json:"hostname"`
	Username string    `json:"username"`
	UID      string    `json:"uid"`
	GID      string    `json:"gid"`
	ShortID  string    `json:"short_id"`
	Paths    []string
}
