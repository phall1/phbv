# phbv dev tasks — run `just` to list, `just <task>` to run.
binary  := "phbv"
sha     := `git rev-parse --short HEAD 2>/dev/null || echo dev`
ldflags := "-s -w -X main.gitSHA=" + sha

# list tasks
default:
    @just --list

# build ./bin/phbv
build:
    go build -ldflags "{{ldflags}}" -o bin/{{binary}} .

# install onto PATH (go env GOBIN, else $GOPATH/bin)
install:
    go install -ldflags "{{ldflags}}" .

# build + run against the cwd's .beads (e.g. `just run --dir foo/.beads`)
run *args:
    go run . {{args}}

# run all tests (set PHBV_TEST_BEADS_DIR for the live smoke test)
test:
    go test ./...

# gofmt + vet + test gate
check:
    @test -z "$(gofmt -l .)" || { echo "gofmt needed:"; gofmt -l .; exit 1; }
    go vet ./...
    go test ./...

# format all Go files
fmt:
    gofmt -w .

# build a local release snapshot (no publish) to validate .goreleaser.yaml
snapshot:
    goreleaser release --snapshot --clean

# tag and push a release — CI builds + publishes (e.g. `just release v0.1.0`)
release version:
    git tag -a {{version}} -m "{{version}}"
    git push origin {{version}}

# remove build artifacts
clean:
    rm -rf bin dist
