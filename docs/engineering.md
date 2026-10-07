# Engineering Rules

- Keep dependencies minimal.
- Validate data at external boundaries.
- Keep authorization in the server.
- Never log credentials, tokens, cookies, or secrets.
- Use context and timeouts for network operations.
- Prefer explicit domain types over ambiguous flags.
- Keep handlers thin.
- Keep database access out of transport code.
- Add tests for failure behavior, not only successful paths.
- Comments should explain non-obvious decisions, not restate code.
