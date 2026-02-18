package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func apiRequest(method, path string, body interface{}) ([]byte, error) {
	url := baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Auto-add admin key for admin paths
	if strings.HasPrefix(path, "/admin/") && adminKey != "" {
		req.Header.Set("X-Admin-Key", adminKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return respBody, nil
}

func apiGet(path string) ([]byte, error) {
	return apiRequest(http.MethodGet, path, nil)
}

func apiPost(path string, body interface{}) ([]byte, error) {
	return apiRequest(http.MethodPost, path, body)
}

func apiPatch(path string, body interface{}) ([]byte, error) {
	return apiRequest(http.MethodPatch, path, body)
}

func apiDelete(path string) ([]byte, error) {
	return apiRequest(http.MethodDelete, path, nil)
}

func apiDeleteWithBody(path string, body interface{}) ([]byte, error) {
	return apiRequest(http.MethodDelete, path, body)
}

// printJSON pretty-prints raw JSON bytes to stdout.
func printJSON(data []byte) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		// Not valid JSON — print raw
		fmt.Println(string(data))
		return
	}
	fmt.Println(buf.String())
}

// printResult prints the API response or exits with error.
func printResult(data []byte, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	printJSON(data)
}
