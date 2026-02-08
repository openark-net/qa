# qa

Federated QA runner for monorepos on unix based systems.

Runs format commands and checks defined in `.qa.yml` files. Checks are cached using git tree hashes—unchanged code is skipped automatically.

In a mono-repo the number of QA checks you have can grow to be unwieldy. You need to format/lint/unit test the front
end, backend, maybe terraform, possibly you have other checks as well. Ideally you should be doing this everytime you
commit. 

## Install

Installs on unix based systems to `/usr/local/bin/qa`
```bash
curl -sL https://raw.githubusercontent.com/openark-net/qa/main/install.sh | bash
```

## Usage

```bash
qa                  # run checks from .qa.yml
qa --no-cache       # run all checks, skip cache
qa init hook        # install pre-commit hook
```

## Configuration

Create a `.qa.yml` in your project root:

```yaml
format:
  - go fmt ./...

checks:
  - go vet ./...
  - go test ./...
```

### Monorepo Setup

Use `includes` to compose checks across subdirectories:

```yaml
includes:
  - api/.qa.yml
  - web/.qa.yml
  - mobile/.qa.yml

format:
  - ./scripts/lint-all
```

Each subdirectory has its own `.qa.yml`:

```yaml
# web/.qa.yml
format:
  - npx prettier --write .

checks:
  - npm run build
  - npm test
```

### Fields

| Field | Description |
|-------|-------------|
| `format` | Commands run sequentially before checks |
| `checks` | Commands run in parallel with caching |
| `includes` | Paths to other `.qa.yml` files |


## Caching

Checks are cached per directory using git tree hashes. A check is skipped when:
1. The directory has no uncommitted changes
2. The git tree hash matches the last successful run

```
$ qa
✓ api: go test ./...       (3.4s)
✓ web: npm test            (5.1s)

# edit api/handler.go, run again
$ qa
✓ api: go test ./...       (2.8s)
○ web: npm test            (cached)
```

Cache is stored in `~/.cache/qa`. Use `--no-cache` to bypass.


## Why?!

Personally I work in a lot of mono repos. I used to have something similar to this as a bashscript that would run
things in parallel but didn't have caching. Running my unit tests in platformio repos can take over 10 seconds, and this
would happen every time I made a commit in a non-platformio part of the repo. 

So I spent an hour or two building this....

How many years of commits will I need to make in order to start saving time?! Who knows?

## Windows

Sorry I don't have a way to test this on Windows. 