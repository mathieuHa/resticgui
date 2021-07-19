dist=${CURDIR}/dist

printdist:
	 echo ${dist}

install:
	cd web && npm install

build: buildgo buildweb

buildgo:
	cd cmd/resticgui && go build -o ${dist}/resticgui.exe

buildweb: install
	cd web && ng build --output-path ${dist}/front/

run:
	cd ${dist}/ && .\resticgui.exe

testweb: install
	cd web && ng test --progress false --watch false

testgo:
	cd cmd/resticgui && go test

clean:
	rm -rf ${dist}/*
