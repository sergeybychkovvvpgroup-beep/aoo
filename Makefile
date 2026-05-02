APP := aoo

.PHONY: build test validate tidy snapshot

build:
	@echo "[build] compiling $(APP)"
	@mkdir -p bin
	@go build -o bin/$(APP) ./cmd/aoo
	@echo "[build] done: bin/$(APP)"

test:
	@echo "[test] running go test"
	@go test ./...
	@echo "[test] done"

validate: build
	@echo "[validate] checking notes in ./notes"
	@./bin/$(APP) validate --dir ./notes
	@echo "[validate] done"

tidy:
	@echo "[tidy] syncing go modules"
	@go mod tidy
	@echo "[tidy] done"

snapshot:
	@echo "[release] building snapshot"
	@goreleaser release --snapshot --clean
