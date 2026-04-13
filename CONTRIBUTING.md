## Dependency management:
No third party dependency. This is a golden rule for this pet project. I want to keep it simple and avoid the hassle of setting up and maintaining Dependabot alerts, haha.

## Error handling rule:
- Use fmt.Fprintln for simple user-facing errors and warnings. Use os.exit(1) to exit with a non-zero status code if it's a fatal error. I recommend calling this from a main function.

- Use fmt.Errorf when you want to return an error value with more context.
`` return fmt.Errorf("read config %q: %w", path, err)``
  - Use %w when callers should be able to inspect the underlying error.
  - Use %v when you only want a human-readable message
- Use errors.Unwrap to check for inspect the underlying error wrapped by fmt.Errorf. 

