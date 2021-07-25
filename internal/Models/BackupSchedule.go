package Models

import "github.com/google/uuid"

type BackupSchedule struct {
	Base
	Name                string `json:"name" gorm:"unique"`
	EveryDay            uint   `json:"every_day"`
	EveryWeek           bool   `json:"every_week"`
	EveryMonth          bool   `json:"every_month"`
	EveryMonday         bool   `json:"every_monday"`
	EveryTuesday        bool   `json:"every_tuesday"`
	EveryWednesday      bool   `json:"every_wednesday"`
	EveryThursday       bool   `json:"every_thursday"`
	EveryFriday         bool   `json:"every_friday"`
	EverySaturday       bool   `json:"every_saturday"`
	EverySunday         bool   `json:"every_sunday"`
	DailyHour           string   `json:"daily_hour"`
	CronSyntax          string   `json:"cron_syntax"`
	RepositoryId        uuid.UUID `json:"repository_id"`
}
