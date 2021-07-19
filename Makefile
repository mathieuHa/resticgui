install:
	cd web && npm install

build:
	cd cmd/resticgui && go build -o ../../dist/resticgui.exe
	cd web && ng build --output-path ../dist/front/

buildgo:
	cd cmd/resticgui && go build -o ../../dist/resticgui.exe

buildfront:
	cd web && ng build --output-path ../dist/front/

run:
	cd dist/ && .\resticgui.exe