package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"sync"
	"time"
)

type requestLogger struct {
	mu  sync.Mutex
	out io.Writer
}

func (l *requestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	timestamp := time.Now().Format(time.RFC3339Nano)
	dump, err := httputil.DumpRequest(r, true)

	// Keep concurrent requests together in the terminal.
	l.mu.Lock()
	var entry bytes.Buffer
	fmt.Fprintln(&entry, timestamp)
	entry.Write(dump)
	if err != nil {
		fmt.Fprintf(&entry, "\nRequest dump error: %v\n", err)
	}
	if !bytes.HasSuffix(entry.Bytes(), []byte("\n")) {
		entry.WriteByte('\n')
	}
	entry.WriteString("\n\n\n")
	_, writeErr := l.out.Write(entry.Bytes())
	l.mu.Unlock()
	if writeErr != nil {
		log.Printf("write request log: %v", writeErr)
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	ip := flag.String("ip", "127.0.0.1", "IP address to listen on (0.0.0.0 for all IPv4 interfaces)")
	port := flag.Int("port", 8080, "TCP port to listen on (0 for an automatically assigned port)")
	flag.Parse()
	if net.ParseIP(*ip) == nil {
		log.Fatalf("invalid listen IP: %q", *ip)
	}
	if *port < 0 || *port > 65535 {
		log.Fatalf("port must be between 0 and 65535")
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(*ip, strconv.Itoa(*port)))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on %s", listener.Addr())
	server := &http.Server{Handler: &requestLogger{out: os.Stdout}}
	log.Fatal(server.Serve(listener))
}
