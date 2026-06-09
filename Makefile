PLUGIN_ID ?= com.wijayacorp.message-forward
PLUGIN_VERSION ?= 0.1.4
BUNDLE_NAME := $(PLUGIN_ID)-$(PLUGIN_VERSION).tar.gz
BUILD_HASH = $(shell git rev-parse HEAD 2>/dev/null || echo "none")
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS += -X "main.buildHash=$(BUILD_HASH)" -X "main.buildDate=$(BUILD_DATE)"
.PHONY: all server webapp bundle clean test
all: server webapp bundle
server:
	mkdir -p server/dist
	cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -trimpath -o dist/plugin-linux-amd64 .
	cd server && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags '$(LDFLAGS)' -trimpath -o dist/plugin-linux-arm64 .
	cd server && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -trimpath -o dist/plugin-darwin-amd64 .
	cd server && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags '$(LDFLAGS)' -trimpath -o dist/plugin-darwin-arm64 .
	cd server && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -trimpath -o dist/plugin-windows-amd64.exe .
webapp:
	cd webapp && npm install --legacy-peer-deps && npm run build
bundle:
	rm -rf dist && mkdir -p dist/$(PLUGIN_ID)/server/dist dist/$(PLUGIN_ID)/webapp/dist
	cp plugin.json dist/$(PLUGIN_ID)/
	cp -r server/dist/* dist/$(PLUGIN_ID)/server/dist/
	cp -r webapp/dist/* dist/$(PLUGIN_ID)/webapp/dist/
	cp -r assets dist/$(PLUGIN_ID)/
	cd dist && tar -czf ../$(BUNDLE_NAME) $(PLUGIN_ID)
clean:
	rm -rf dist server/dist webapp/dist webapp/node_modules *.tar.gz
test:
	cd server && go test ./... -v
	cd webapp && npm run build
