//go:build ignore

// gen_fixture writes legacy-v0.1.10.db, a database in the format of the
// ent-based ResticGUI (v0.1.10), used by the legacy import tests.
//
//	go run internal/legacy/testdata/gen_fixture.go
//
// It needs the ent code, which was removed after commit 0e49007: check that
// commit out to run it. The generated file is committed.
package main

import (
	"context"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mathieuHa/resticgui/ent"
)

const out = "internal/legacy/testdata/legacy-v0.1.10.db"

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func main() {
	ctx := context.Background()
	os.Remove(out)
	c := must(ent.Open("sqlite3", out+"?_fk=1"))
	defer c.Close()
	if err := c.Schema.Create(ctx); err != nil {
		panic(err)
	}

	// local repository, weekly backups, forget/prune on a cron spec
	docs := must(c.Repository.Create().SetName("local-docs").SetPassword("docs-password").
		SetProvider("localProvider").SetActive(true).SetPruneActive(true).SetInitiated(true).
		SetMaxSize("2G").Save(ctx))
	must(c.LocalProvider.Create().SetRepoPath("/srv/restic/docs").SetRepository(docs).Save(ctx))
	must(c.BackupSchedule.Create().SetChoice("weekly").SetEveryMonday(true).SetEveryFriday(true).
		SetHour("02:30").SetRepository(docs).Save(ctx))
	must(c.PruneSchedule.Create().SetChoice("cron").SetCronSyntax("0 4 * * 0").SetRepository(docs).Save(ctx))
	must(c.BackupPolicy.Create().SetNbKeepLast(5).SetNbKeepHours(0).SetNbKeepDays(7).
		SetNbKeepWeeks(4).SetNbKeepMonths(6).SetRepository(docs).Save(ctx))
	must(c.Path.Create().SetName("docs").SetAbsPath("/home/me/docs").SetRelPath("/home/me/docs").
		SetIsDir(true).SetIncluded(true).SetRepository(docs).Save(ctx))
	must(c.Path.Create().SetName("tmp").SetAbsPath("/home/me/docs/tmp").SetRelPath("/home/me/docs/tmp").
		SetIsDir(true).SetIncluded(false).SetRepository(docs).Save(ctx))
	for _, op := range []string{"GetSnapshots", "BackupRepository"} {
		must(c.Logs.Create().SetOperation(op).SetError("").SetStderr("").SetStdout("{}").
			SetSuccess(true).SetStartTime(time.Date(2023, 6, 1, 2, 30, 0, 0, time.UTC)).SetRepository(docs).Save(ctx))
	}

	// AWS S3, interval written the way the UI placeholder suggests ("1h30"), backups disabled
	photos := must(c.Repository.Create().SetName("aws-photos").SetPassword("photos-password").
		SetProvider("AWSS3Provider").SetActive(false).Save(ctx))
	must(c.AWSS3Provider.Create().SetBucketName("my-photos").SetAccessKeyID("AKIAEXAMPLE").
		SetSecretAccessKey("aws-secret-key").SetDefaultRegion("eu-west-3").SetRepository(photos).Save(ctx))
	must(c.BackupSchedule.Create().SetChoice("interval").SetInterval("1h30").SetRepository(photos).Save(ctx))
	must(c.BackupPolicy.Create().SetRepository(photos).Save(ctx)) // ent defaults: last 10, 1 of each
	must(c.Path.Create().SetName("pictures").SetAbsPath("/home/me/pictures").SetRelPath("").
		SetIsDir(true).SetRepository(photos).Save(ctx))

	// B2, daily cron
	archive := must(c.Repository.Create().SetName("b2-archive").SetPassword("archive-password").
		SetProvider("B2Provider").SetActive(true).Save(ctx))
	must(c.B2Provider.Create().SetBucketName("archive").SetRepoName("laptop").SetAccountID("b2-account").
		SetAccountKey("b2-secret-key").SetRepository(archive).Save(ctx))
	must(c.BackupSchedule.Create().SetChoice("cron").SetCronSyntax("15 1 * * *").SetRepository(archive).Save(ctx))

	// MinIO with a custom CA, monthly schedule, and an invalid prune cron
	lab := must(c.Repository.Create().SetName("minio-lab").SetPassword("lab-password").
		SetProvider("MinioS3Provider").SetActive(true).SetPruneActive(true).Save(ctx))
	must(c.MinioS3Provider.Create().SetBucketName("lab").SetHostname("10.0.0.5").SetPort("9000").
		SetScheme("https").SetAccessKeyID("minio-access").SetSecretAccessKey("minio-secret-key").
		SetCertificateAuthority("/etc/ssl/minio-ca.pem").SetRepository(lab).Save(ctx))
	must(c.BackupSchedule.Create().SetChoice("monthly").SetEveryMonth(true).SetDay("1").SetHour("03:00").
		SetRepository(lab).Save(ctx))
	must(c.PruneSchedule.Create().SetChoice("cron").SetCronSyntax("every sunday").SetRepository(lab).Save(ctx))

	// created but never configured
	must(c.Repository.Create().SetName("unconfigured").SetPassword("lonely-password").Save(ctx))
}
