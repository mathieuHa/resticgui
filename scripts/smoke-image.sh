#!/bin/sh
# Smoke test of a built image: it starts, serves the UI and the API, and
# backs up a folder to a local repository with the bundled restic.
# Usage: scripts/smoke-image.sh IMAGE [PLATFORM] [VERSION]
set -eu

image=$1
platform=${2:-}
version=${3:-}
name=resticgui-smoke-$$
work=$(mktemp -d)

cleanup() {
	status=$?
	if [ "$status" -ne 0 ]; then docker logs "$name" 2>&1 | tail -50 || true; fi
	docker rm -f "$name" >/dev/null 2>&1 || true
	docker run --rm -v "$work:/w" alpine:3.24 rm -rf /w/data /w/repo >/dev/null 2>&1 || true
	rm -rf "$work"
	exit "$status"
}
trap cleanup EXIT

mkdir -p "$work/data/docs"
echo "hello from the smoke test" > "$work/data/docs/hello.txt"

docker run -d --name "$name" ${platform:+--platform "$platform"} \
	-p 127.0.0.1::3030 -v "$work:/backup" "$image" >/dev/null
base="http://$(docker port "$name" 3030/tcp | head -1)"
api="$base/api/v2"

# emulated platforms start slowly
i=0
until curl -fs -o /dev/null "$base/ui/"; do
	i=$((i + 1))
	[ "$i" -lt 120 ] || { echo "the UI did not answer"; exit 1; }
	sleep 1
done
echo "UI is up at $base/ui/"

curl -fsS "$api/auth/session" | jq -e '.authenticated == true' >/dev/null
got=$(curl -fsS "$api/version" | jq -r .version)
echo "version: $got"
[ -z "$version" ] || [ "$got" = "$version" ] || { echo "expected version $version"; exit 1; }
docker exec "$name" restic version

post() { curl -fsS -X "$1" -H 'Content-Type: application/json' -d "$3" "$api$2"; }

id=$(post POST /repositories '{"name":"smoke","backend":"local","location":"/backup/repo","password":"smoke test password"}' | jq -r .id)
post POST "/repositories/$id/plans" '{"name":"Default","paths":{"include":["/backup/data"],"exclude":[]}}' >/dev/null

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
echo "smoke test passed"
