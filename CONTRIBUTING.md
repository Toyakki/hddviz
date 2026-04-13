# Contribution guide

## Non-technical guidelines
By participating in this project, you agree to abide by the [code of conduct](https://github.com/goreleaser/.github/blob/main/CODE_OF_CONDUCT.md). 

## Personal note
This is my first golang pet project, and I am still learning this language and how to deploy it properly. Any constructive criticism and suggestions are very welcome.

## Setup
* Go 1.25+

Clone this repository and navigate to the project directory. You can run the CLI tool by building it using `go build .` and then execute it with `go run .`.

This is optional, but I highly recommend installing golangci-lint locally before pushing any code. My GHA uses this linter to check for code quality issues.

## Technical guidelines
### 1. Dependency management
**No third party dependency for the application logic.** This is a golden rule for this pet project. I want to keep it simple and avoid the hassle of setting up and maintaining Dependabot alerts, haha.

### 2. Error handling rule
Here are my custom rules that I decided after conversing with GPT and reading some articles on Go error handling. Feel free to suggest any changes or improvements to these rules.
- Use fmt.Fprintln for simple user-facing errors and warnings. Use os.exit(1) to exit with a non-zero status code if it's a fatal error. I recommend calling this from a main function.
- Use fmt.Errorf when you want to return an error value with more context.
`` return fmt.Errorf("read config %q: %w", path, err)``
  - Use %w when callers should be able to inspect the underlying error.
  - Use %v when you only want a human-readable message
- Use errors.Unwrap to check to inspect the underlying error wrapped by fmt.Errorf. 

### Good references
https://github.com/uber-go/guide/tree/master
https://www.datadoghq.com/blog/go-error-handling/

That's all for now!