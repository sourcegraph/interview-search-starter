# Interview Starter: Single-repo Code Search

A minimal Go HTTP service that searches a single Git repository using `git grep`.

## Usage

```bash
go run main.go --root ~/sourcegraph --port 8080
```

**Flags:**
- `--root` — Directory containing Git repositories (default: `.`)
- `--port` — Port to listen on (default: `8080`)

## API

### `GET /search`

Search a single repository.

**Parameters:**
- `query` — The search pattern (passed to `git grep`)
- `repo` — The repository name (subdirectory under `--root`)

**Example:**

```bash
curl "http://localhost:8080/search?query=func+main&repo=sourcegraph"
```

**Response:**

```json
[
  {
    "file": "cmd/server/main.go",
    "line": 42,
    "content": "func main() {",
    "score": 0.69
  }
]
```

Results are scored by match density: shorter lines where the query covers more of the content score higher (range 0–1).
