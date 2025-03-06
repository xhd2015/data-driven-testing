package svg

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/xhd2015/data-driven-testing/decision_tree"
)

// Server serves SVG content dynamically
type Server struct {
	renderer *Renderer
	srv      *http.Server

	// content source
	source   serverSource        // indicates whether serving from memory or file
	tree     *decision_tree.Node // in-memory tree
	filename string              // file path when serving from file
	mu       sync.RWMutex        // protects tree and filename

	// shutdown
	shutdownCh chan struct{}
	doneCh     chan struct{}

	// testing
	portCh chan<- int
}

type serverSource int

const (
	sourceMemory serverSource = iota
	sourceFile
)

// NewServer creates a new SVG server with the given renderer
func NewServer(renderer *Renderer) *Server {
	if renderer == nil {
		renderer = NewRenderer(decision_tree.DefaultConfig())
	}
	return &Server{
		renderer:   renderer,
		shutdownCh: make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

// SetPortNotifier sets a channel to receive the port number when the server starts.
// This is primarily used for testing.
func (s *Server) SetPortNotifier(ch chan<- int) {
	s.portCh = ch
}

// tryListen attempts to listen on the given port
func tryListen(port int) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf(":%d", port))
}

// Serve starts serving the SVG content for the given tree.
// It automatically selects an available port starting from 12137,
// and prints the URL to console.
// The server runs until Stop is called or the process receives SIGINT/SIGTERM.
func (s *Server) Serve(tree *decision_tree.Node) error {
	s.mu.Lock()
	s.source = sourceMemory
	s.tree = tree
	s.filename = "" // clear filename if any
	s.mu.Unlock()

	return s.startServer()
}

// ServeFile starts serving SVG content from the given JSON file.
// The file should contain a JSON representation of a decision_tree.Node.
// The server will read the file on each request to ensure latest content.
func (s *Server) ServeFile(filename string) error {
	// Verify file exists and is readable
	if _, err := os.Stat(filename); err != nil {
		return fmt.Errorf("check file: %v", err)
	}

	s.mu.Lock()
	s.source = sourceFile
	s.filename = filename
	s.tree = nil // clear any in-memory tree
	s.mu.Unlock()

	return s.startServer()
}

// startServer starts the HTTP server and handles shutdown
func (s *Server) startServer() error {
	// Try ports starting from defaultPort
	var listener net.Listener
	var err error
	var port int

	// Try a range of ports
	for port = defaultPort; port < defaultPort+10; port++ {
		listener, err = tryListen(port)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("failed to find available port in range %d-%d: %v", defaultPort, port-1, err)
	}

	// Create server
	s.srv = &http.Server{
		Handler: s.handler(),
	}

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Print server URL
	fmt.Printf("SVG server running at: http://localhost:%d\n", port)

	// Notify port for testing if channel is set
	if s.portCh != nil {
		s.portCh <- port
	}

	// Start server in goroutine
	go func() {
		defer close(s.doneCh)
		if err := s.srv.Serve(listener); err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	// Wait for shutdown signal
	select {
	case <-sigCh:
		fmt.Println("\nShutting down server...")
	case <-s.shutdownCh:
		fmt.Println("Server stop requested...")
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	return s.srv.Shutdown(ctx)
}

// UpdateTree updates the tree being served (when in memory mode)
func (s *Server) UpdateTree(tree *decision_tree.Node) {
	s.mu.Lock()
	if s.source == sourceMemory {
		s.tree = tree
	}
	s.mu.Unlock()
}

// UpdateFile updates the file path being served (when in file mode)
func (s *Server) UpdateFile(filename string) error {
	if _, err := os.Stat(filename); err != nil {
		return fmt.Errorf("check file: %v", err)
	}
	s.mu.Lock()
	if s.source == sourceFile {
		s.filename = filename
	}
	s.mu.Unlock()
	return nil
}

// Stop stops the server gracefully
func (s *Server) Stop() error {
	if s.srv == nil {
		return nil
	}
	close(s.shutdownCh)
	<-s.doneCh // wait for server to finish
	return nil
}

// handler creates the HTTP handler for serving SVG content
func (s *Server) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		source := s.source
		tree := s.tree
		filename := s.filename
		s.mu.RUnlock()

		if source == sourceFile {
			if filename == "" {
				http.Error(w, "No file configured", http.StatusNotFound)
				return
			}

			// Read and parse file
			jsonData, err := os.ReadFile(filename)
			if err != nil {
				http.Error(w, fmt.Sprintf("read file: %v", err), http.StatusInternalServerError)
				return
			}

			if err := json.Unmarshal(jsonData, &tree); err != nil {
				http.Error(w, fmt.Sprintf("parse JSON: %v", err), http.StatusInternalServerError)
				return
			}
		} else if tree == nil {
			http.Error(w, "No tree available", http.StatusNotFound)
			return
		}

		// Set headers
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Generate and write SVG
		svg := s.renderer.RenderTree(tree)
		if _, err := w.Write([]byte(svg)); err != nil {
			fmt.Fprintf(os.Stderr, "error writing response: %v\n", err)
		}
	})
}

// defaultShutdownTimeout is the time to wait for server to shutdown gracefully
const defaultShutdownTimeout = 5 * time.Second

// defaultPort is the starting port number to try
const defaultPort = 12137
