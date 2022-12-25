package Models

import "time"

type Files struct {
	Time       time.Time `json:"time"`
	Parent     string    `json:"parent"`
	Tree       string    `json:"tree"`
	Hostname   string    `json:"hostname"`
	Username   string    `json:"username"`
	UID        int       `json:"uid"`
	GID        int       `json:"gid"`
	ID         string    `json:"id"`
	StructType string    `json:"struct_type"`
	ShortID    string    `json:"short_id"`
	Path       string    `json:"path"`
	Type       string    `json:"type"`
	Mode       int       `json:"mode"`
	Size       int       `json:"size"`
	Mtime      string    `json:"mtime"`
}
