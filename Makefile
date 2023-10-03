CMDS := .build\bin\houndd .build\bin\hound

SRCS := $(shell dir /s /b *.go)
UI := $(shell dir /s /b ui\assets)

WEBPACK_ARGS := --mode production
ifdef DEBUG
	WEBPACK_ARGS := --mode development
endif

ALL: $(CMDS)

ui: ui\.build\ui

node_modules\build:
	npm install
	echo %date% %time% >> $@

.build\bin\houndd: ui\.build\ui $(SRCS)
	go build -o $@ ./cmds/houndd

.build\bin\hound: $(SRCS)
	go build -o $@ ./cmds/hound

ui\.build\ui: node_modules\build $(UI)
	if not exist ui\.build\ui mkdir ui\.build\ui
	xcopy /s /y /exclude:exclude.txt ui\assets ui\.build\ui
	npx webpack $(WEBPACK_ARGS)

dev: node_modules\build

test:
	go test ./...
	npm test

lint:
	set GO111MODULE=on
	go get github.com\golangci\golangci-lint\cmd\golangci-lint
	golangci-lint run ./...

clean:
	rmdir /s /q .build ui\.build node_modules
