# Interview Starter: Single-repo Code Search

A minimal Go HTTP service that searches a single Git repository using `git grep`.

It is the starting point for a coding exercise interview in which the candidate will build out features on top of this service.

The candidate is required to use an agentic coding assistant (Claude Code, Amp, Cursor, Codex, etc...) during the interview.

## Usage

```bash
go run main.go --root ~/src --port 48080
```

**Flags:**
- `--root` — Directory containing Git repositories (default: `.`)
- `--port` — Port to listen on (default: `48080`)

## API

### `GET /search`

Search a single repository.

**Parameters:**
- `query` — The search pattern (passed to `git grep`)
- `repo` — The repository name (subdirectory under `--root`)

**Example:**

```bash
curl "http://localhost:48080/search?query=func+main&repo=myrepo"
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

## Setup (before the interview)

### 1. Clone the starter repo and set up the workspace

```bash
mkdir -p ~/interview-search/repos
git clone https://github.com/sourcegraph/interview-search-starter.git ~/interview-search/interview-search-starter
```

### 2. Clone OSS repositories to search against

These are a mix of languages and sizes:

```bash
cd ~/interview-search/repos
git clone https://github.com/go-chi/chi               # Go - HTTP router
git clone https://github.com/pallets/flask            # Python - web framework
git clone https://github.com/expressjs/express        # JavaScript - web framework
git clone https://github.com/serde-rs/serde           # Rust - serialization
git clone https://github.com/jqlang/jq                # C - JSON processor
```

### 3. Start the server

```bash
cd ~/interview-search/interview-search-starter
go run main.go --root ~/interview-search/repos --port 48080
```

### 4. Verify it works

```bash
curl 'http://localhost:48080/search?query=func+main&repo=chi'
```

Or open http://localhost:48080/ in a browser to use the web UI.

### 5. Ensure your coding assistant is ready to go

This exercise requires an agentic coding assistant (Claude Code, Amp, Cursor, Codex, etc...), so get yours ready to go with this repo.
