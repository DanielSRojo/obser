# List available recipes
default:
    @just --list

# Build the obser binary
build:
    go build -o obser .

# Install obser into $GOBIN (or $GOPATH/bin)
install:
    go install .

# Run obser with the given arguments, e.g. `just run journal entries`
run *args:
    go run . {{args}}

# Run the tests
test *args:
    go test ./... {{args}}

# Format the code with gofumpt
fmt:
    gofumpt -w .

# Fail if any file is not formatted with gofumpt
fmt-check:
    test -z "$(gofumpt -l .)" || { gofumpt -l .; exit 1; }

# Run go vet and golangci-lint
lint:
    go vet ./...
    golangci-lint run ./...

# Run every check: formatting, lint and tests
check: fmt-check lint test

# Tidy go.mod and go.sum
tidy:
    go mod tidy

# Remove build output
clean:
    rm -f obser

# obser commands, run from source against $OBSER_VAULT (or pass --vault <path>)

# List the journal entries
[group('obser')]
journal-entries *args:
    go run . journal entries {{args}}

# Create missing journal entries from the template (--dry-run, --from, --to)
[group('obser')]
journal-fill *args:
    go run . journal fill {{args}}

# Remove empty properties from past journal entries (--dry-run, --before)
[group('obser')]
journal-tidy *args:
    go run . journal tidy {{args}}

# List the notes in the vault
[group('obser')]
list-notes *args:
    go run . list notes {{args}}

# List the properties used in notes
[group('obser')]
list-properties *args:
    go run . list properties {{args}}

# Show how consistently each habit is done (--year, --month)
[group('obser')]
habits *args:
    go run . habits {{args}}

# Show one habit's calendar and breakdown for a year (--year)
[group('obser')]
habits-show habit *args:
    go run . habits show {{quote(habit)}} {{args}}

# Compare habits across months or years (--by, --metric, --year)
[group('obser')]
habits-trend *args:
    go run . habits trend {{args}}

# Create missing journal entries, then remove their empty properties (--dry-run)
[group('obser')]
fill *args:
    go run . journal fill {{args}}
    go run . journal tidy {{args}}
