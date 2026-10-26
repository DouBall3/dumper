package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRequestLogging(t *testing.T) {
	for _, method := range []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "CUSTOM"} {
		t.Run(method, func(t *testing.T) {
			var output bytes.Buffer
			handler := &requestLogger{out: &output}
			request := httptest.NewRequest(method, "/example?key=value", strings.NewReader("hello\nworld"))
			request.Header.Set("X-Example", "test-value")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusOK || response.Body.Len() != 0 {
				t.Fatalf("response = %d %q; want 200 with empty body", response.Code, response.Body.String())
			}
			stamp, dump, ok := strings.Cut(output.String(), "\n")
			if !ok {
				t.Fatal("missing timestamp line")
			}
			if _, err := time.Parse(time.RFC3339Nano, stamp); err != nil {
				t.Fatalf("invalid timestamp: %v", err)
			}
			for _, expected := range []string{method + " /example?key=value HTTP/1.1\r\n", "X-Example: test-value\r\n", "hello\nworld\n\n\n\n"} {
				if !strings.Contains(dump, expected) {
					t.Errorf("log missing %q: %q", expected, dump)
				}
			}
		})
	}
}

func TestConcurrentRequests(t *testing.T) {
	var output bytes.Buffer
	handler := &requestLogger{out: &output}
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			request := httptest.NewRequest("POST", fmt.Sprintf("/request/%d", i), strings.NewReader(fmt.Sprintf("body-%d", i)))
			handler.ServeHTTP(httptest.NewRecorder(), request)
		}(i)
	}
	wg.Wait()
	entries := strings.Split(strings.TrimSuffix(output.String(), "\n\n\n\n"), "\n\n\n\n")
	if len(entries) != 30 {
		t.Fatalf("got %d log entries, want 30", len(entries))
	}
	for _, entry := range entries {
		if strings.Count(entry, "POST /request/") != 1 || strings.Count(entry, "body-") != 1 {
			t.Fatalf("interleaved or incomplete entry: %q", entry)
		}
	}
}

func TestCORS(t *testing.T) {
	for _, method := range []string{http.MethodOptions, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			var output bytes.Buffer
			handler := &requestLogger{out: &output}
			request := httptest.NewRequest(method, "/", nil)
			request.Header.Set("Origin", "http://localhost:5173")
			if method == http.MethodOptions {
				request.Header.Set("Access-Control-Request-Method", "POST")
				request.Header.Set("Access-Control-Request-Headers", "content-type,authorization")
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			result := response.Result()
			defer result.Body.Close()
			if result.StatusCode != http.StatusOK {
				t.Fatalf("status = %d; want 200", result.StatusCode)
			}
			for name, want := range map[string]string{
				"Access-Control-Allow-Origin":  "http://localhost:5173",
				"Access-Control-Allow-Methods": "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			} {
				if got := result.Header.Get(name); got != want {
					t.Errorf("%s = %q; want %q", name, got, want)
				}
			}
			if !strings.Contains(output.String(), method+" / HTTP/1.1") {
				t.Error("request was not logged")
			}
		})
	}
}
