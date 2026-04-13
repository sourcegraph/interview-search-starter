package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Result struct {
	File    string  `json:"file"`
	Line    int     `json:"line"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

var rootDir string

func main() {
	flag.StringVar(&rootDir, "root", ".", "Root directory containing git repositories")
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	var err error
	rootDir, err = filepath.Abs(rootDir)
	if err != nil {
		log.Fatalf("invalid root directory: %v", err)
	}

	http.HandleFunc("/search", handleSearch)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Listening on %s (root: %s)", addr, rootDir)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	repo := r.URL.Query().Get("repo")

	if query == "" || repo == "" {
		http.Error(w, `{"error": "query and repo parameters are required"}`, http.StatusBadRequest)
		return
	}

	repoPath, err := resolveRepoPath(rootDir, repo)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": %q}`, err.Error()), http.StatusBadRequest)
		return
	}

	results, err := gitGrep(repoPath, query)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": %q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func resolveRepoPath(root, repo string) (string, error) {
	if repo == "" || filepath.IsAbs(repo) {
		return "", fmt.Errorf("invalid repo name")
	}

	cleaned := filepath.Clean(repo)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid repo name")
	}

	repoPath := filepath.Join(root, cleaned)
	rel, err := filepath.Rel(root, repoPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid repo name")
	}

	return repoPath, nil
}

func gitGrep(repoPath, query string) ([]Result, error) {
	cmd := exec.Command("git", "grep", "-n", "--no-color", "-I", "-F", "--", query)
	cmd.Dir = repoPath

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("git grep failed: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("git grep failed: %w", err)
	}

	var results []Result
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		if r, ok := parseLine(scanner.Text(), query); ok {
			results = append(results, r)
		}
	}
	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return nil, fmt.Errorf("reading git grep output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []Result{}, nil
		}
		return nil, fmt.Errorf("git grep failed: %w", err)
	}

	return results, nil
}

func parseLine(line, query string) (Result, bool) {
	// git grep -n output format: "file:linenum:content"
	firstColon := strings.Index(line, ":")
	if firstColon < 0 {
		return Result{}, false
	}
	rest := line[firstColon+1:]
	secondColon := strings.Index(rest, ":")
	if secondColon < 0 {
		return Result{}, false
	}

	file := line[:firstColon]
	lineNum, err := strconv.Atoi(rest[:secondColon])
	if err != nil {
		return Result{}, false
	}
	content := rest[secondColon+1:]

	// Simple relevance score: shorter lines with matches score higher,
	// since the match is a larger proportion of the content.
	score := 1.0
	if len(content) > 0 {
		score = float64(len(query)) / float64(len(content))
	}
	if score > 1.0 {
		score = 1.0
	}

	return Result{
		File:    file,
		Line:    lineNum,
		Content: strings.TrimSpace(content),
		Score:   score,
	}, true
}
