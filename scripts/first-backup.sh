#!/bin/sh
# The end of every quick start: with a running ResticGUI, add a repository,
# back a folder up into it and check the snapshot, through the API the web
# interface uses. Usage: scripts/first-backup.sh URL SOURCE REPOSITORY
#   URL         where ResticGUI answers, such as http://127.0.0.1:3001
#   SOURCE      a folder to back up, as ResticGUI sees it
#   REPOSITORY  a local path for the new repository, as ResticGUI sees it
set -eu

base=$1
source=$2
repository=$3
api="$base/api/v2"

i=0
until curl -fs -o /dev/null "$base/ui/"; do
	i=$((i + 1))
	[ "$i" -lt 120 ] || { echo "ResticGUI did not answer at $base"; exit 1; }
	sleep 1
done
echo "ResticGUI $(curl -fsS "$api/version" | jq -r .version) answers at $base/ui/"

post() { curl -fsS -X "$1" -H 'Content-Type: application/json' -d "$3" "$api$2"; }

id=$(post POST /repositories "{\"name\":\"Quick start\",\"backend\":\"local\",\"location\":\"$repository\",\"password\":\"quick start password\"}" | jq -r .id)
post PUT "/repositories/$id/paths" "{\"include\":[\"$source\"],\"exclude\":[]}" >/dev/null
post PUT "/repositories/$id/schedules/backup" '{"cron":"30 2 * * *","enabled":true}' >/dev/null

run() {
	job=$(post POST "/repositories/$id/jobs" "{\"kind\":\"$1\"}" | jq -r .id)
	i=0
	while :; do
		status=$(curl -fsS "$api/jobs/$job" | jq -r .status)
		case "$status" in
		succeeded) echo "$1: succeeded"; return 0 ;;
		queued | running) ;;
		*) echo "$1: $status"; curl -fsS "$api/jobs/$job"; exit 1 ;;
		esac
		i=$((i + 1))
		[ "$i" -lt 300 ] || { echo "$1 did not finish"; exit 1; }
		sleep 1
	done
}
run init
run backup

count=$(curl -fsS "$api/repositories/$id/snapshots" | jq length)
[ "$count" -eq 1 ] || { echo "expected 1 snapshot, got $count"; exit 1; }
echo "first backup done: 1 snapshot of $source in $repository"
