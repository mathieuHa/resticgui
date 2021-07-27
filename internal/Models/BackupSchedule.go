package Models

import "github.com/google/uuid"

type BackupSchedule struct {
	Base
	EveryDay            bool   `json:"every_day"`
	EveryWeek           bool   `json:"every_week"`
	EveryMonth          bool   `json:"every_month"`
	EveryMonday         bool   `json:"every_monday"`
	EveryTuesday        bool   `json:"every_tuesday"`
	EveryWednesday      bool   `json:"every_wednesday"`
	EveryThursday       bool   `json:"every_thursday"`
	EveryFriday         bool   `json:"every_friday"`
	EverySaturday       bool   `json:"every_saturday"`
	EverySunday         bool   `json:"every_sunday"`
	Hour           		string   	`json:"hour"`
	Day           		string   	`json:"day"`
	Interval        	string 		`json:"interval"`
	CronSyntax          string   	`json:"cron_syntax"`
	Choice              string   	`json:"choice"`
	RepositoryID        uuid.UUID	`json:"repository_id"`
}
