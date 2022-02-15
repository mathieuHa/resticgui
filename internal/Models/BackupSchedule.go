package Models

import (
	"github.com/robfig/cron/v3"
)

type BackupSchedule struct {
	Base

	EveryDay       bool         `json:"everyDay"`
	EveryWeek      bool         `json:"everyWeek"`
	EveryMonth     bool         `json:"everyMonth"`
	EveryMonday    bool         `json:"everyMonday"`
	EveryTuesday   bool         `json:"everyTuesday"`
	EveryWednesday bool         `json:"everyWednesday"`
	EveryThursday  bool         `json:"everyThursday"`
	EveryFriday    bool         `json:"everyTriday"`
	EverySaturday  bool         `json:"everySaturday"`
	EverySunday    bool         `json:"everySunday"`
	Hour           string       `json:"hour"`
	Day            string       `json:"day"`
	Interval       string       `json:"interval"`
	CronSyntax     string       `json:"cronSyntax"`
	Choice         string       `json:"choice"`
	RepositoryID   uint         `json:"repositoryID"`
	CronID         cron.EntryID `json:"cronID"`
}
