package backup

import (
	"errors"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/database"
)

func validSchedule(schedule *ent.BackupSchedule) (valid bool, err error) {
	log.Debug().Msg("validSchedule()")

	if schedule == nil {
		valid = false
		err = errors.New("schedule is nil")
		return valid, err
	}
	switch schedule.Choice {
	// check cron
	case "cron":
		return true, nil
		// check interval
	case "interval":
		return true, nil
		// check regular
	case "regular":
		return true, nil
		// check weekly
	case "weekly":
		return true, nil
	default:
		valid = false
		err = errors.New("provider does not exist")
	}
	return valid, err
}

// This returns the schedule in string format that can be used by cron
func getScheduleStringSyntax(schedule ent.BackupSchedule) string {
	log.Debug().Msg("getScheduleStringSyntax()")

	if schedule.Choice == "cron" {
		return schedule.CronSyntax
	} else if schedule.Choice == "interval" {
		return fmt.Sprintf("@every %v", schedule.Interval)
	} else if schedule.Choice == "regular" {
		hour, _ := time.Parse("15:04", schedule.Hour)
		log.Debug().Msgf("hour = %v", hour)
		if schedule.EveryDay {
			return fmt.Sprintf("%d %d * * *", hour.Minute(), hour.Hour())
		} else if schedule.EveryWeek {
			return fmt.Sprintf("%d %d * * %v", hour.Minute(), hour.Hour(), schedule.Day)
		} else if schedule.EveryMonth {
			return fmt.Sprintf("%d %d %v * *", hour.Minute(), hour.Hour(), schedule.Day)
		}
	} else if schedule.Choice == "weekly" {
		// check weekly
		hour, err := time.Parse("15:04", schedule.Hour)
		log.Debug().Err(err).Msgf("hour = %v", hour)
		day := ""
		if schedule.EveryMonday {
			day = fmt.Sprintf("%v1,", day)
		}
		if schedule.EveryTuesday {
			day = fmt.Sprintf("%v2,", day)
		}
		if schedule.EveryWednesday {
			day = fmt.Sprintf("%v3,", day)
		}
		if schedule.EveryThursday {
			day = fmt.Sprintf("%v4,", day)
		}
		if schedule.EveryFriday {
			day = fmt.Sprintf("%v5,", day)
		}
		if schedule.EverySaturday {
			day = fmt.Sprintf("%v6,", day)
		}
		if schedule.EverySunday {
			day = fmt.Sprintf("%v0,", day)
		}
		if len(day) > 0 {
			log.Debug().Msgf("DAY = %v", day)
			day = day[:len(day)-1]
			log.Debug().Msgf("DAY = %v", day)

		}
		return fmt.Sprintf("%d %d * * %v", hour.Minute(), hour.Hour(), day)
	}
	return ""
}

func AddScheduledBackup(repo *ent.Repository, cr *cron.Cron) bool {
	log.Debug().Msg("addScheduledBackup()")

	// we check if the cron has been initialized
	if cr == nil {
		log.Error().Msg("cr is not initialied")
		return false
	}

	// we check if the repository is valid
	if !repo.Valid {
		log.Debug().Msg("repository is not valid")
		return false
	} else {
		log.Debug().Msg("repository is valid")

		// First we remove the previous job launched for this repository
		log.Debug().Msg("remove previous job for the repository")
		if repo.Edges.BackupSchedule != nil {
			if repo.Edges.BackupSchedule.CronID != nil {
				cr.Remove(cron.EntryID(*repo.Edges.BackupSchedule.CronID))
			}
			// Then we had the new job to be run and register it
			timing := getScheduleStringSyntax(*repo.Edges.BackupSchedule)
			cronID, err := cr.AddFunc(timing, func() {
				BackupRepository(repo)
			})

			if err != nil {
				log.Error().Err(err)
				return false
			}

			err = database.UpdateCronID(*repo.Edges.BackupSchedule, int(cronID))
			if err != nil {
				log.Err(err).Msgf("error: repository %v", "")
				return false
			}
		}
		return true
	}

}

func CreateAllScheduledBackup(cr *cron.Cron) {
	log.Debug().Msg("CreateAllScheduledBackup()")

	// Check the validity of the provider and repository in general
	repositories, err := database.QueryActiveRWithBackupSchedule()
	if err != nil {
		log.Err(err).Msgf("error quering active backups")
	} else {
		for _, repo := range repositories {
			log.Debug().Msgf("start repos %v", repo.ID)
			AddScheduledBackup(repo, cr)
		}
	}
}

func RemoveAllScheduledBackup(cr *cron.Cron) {
	log.Debug().Msg("RemoveAllScheduledBackup()")

	// Check the validity of the provider and repository in general
	repositories, err := database.QueryAllRWithBackupSchedule()
	if err != nil {
		log.Err(err).Msgf("error quering active backups")
	} else {
		for _, repo := range repositories {
			log.Debug().Msgf("start repos %v", repo.ID)
			removeScheduledBackup(repo, cr)
		}
	}

}

func removeScheduledBackup(repository *ent.Repository, cr *cron.Cron) {
	log.Debug().Msg("removeScheduledBackup()")

	if repository == nil {
		return
	}
	// Here we need to
	if cr != nil {
		log.Debug().Msg("Removing backups")
		// First we remove the previous job launched for this repository
		if repository.Edges.BackupSchedule != nil && repository.Edges.BackupSchedule.CronID != nil {
			log.Debug().Msg("repository.Edges.BackupSchedule != nil && *repository.Edges.BackupSchedule.CronID > 0")

			id := cron.EntryID(*repository.Edges.BackupSchedule.CronID)
			cr.Remove(id)
			database.UpdateCronID(*repository.Edges.BackupSchedule, 0)
		}
	}
}
