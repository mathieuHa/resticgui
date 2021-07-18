install:
	echo "test"

build:
	cd restic && go build -o ../dist/resticgui.exe
	cd web && ng build --output-path ../dist/front/

buildgo:
	cd restic && go build -o ../dist/resticgui.exe

buildfront:
	cd web && ng build --output-path ../dist/front/

run:
	cd dist/ && .\resticgui.exe