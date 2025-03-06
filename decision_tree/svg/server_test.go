package svg

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/data-driven-testing/decision_tree"
)

func TestServer(t *testing.T) {
	// Create a simple test tree
	tree := &decision_tree.Node{
		ID:    "root",
		Label: "Root Node",
		Children: []*decision_tree.Node{
			{
				ID:    "child1",
				Label: "Child 1",
			},
		},
	}

	t.Run("PortSelection", func(t *testing.T) {
		// Try to find a free port in our range
		var firstFreePort int
		for p := defaultPort; p < defaultPort+10; p++ {
			if l, err := net.Listen("tcp", fmt.Sprintf(":%d", p)); err == nil {
				l.Close()
				firstFreePort = p
				break
			}
		}
		if firstFreePort == 0 {
			t.Skip("no free ports available in range")
		}

		// Create server
		server := NewServer(nil)

		// Channel to receive the port
		portCh := make(chan int, 1)
		server.SetPortNotifier(portCh)

		// Start server in background
		go func() {
			if err := server.Serve(tree); err != nil && err != http.ErrServerClosed {
				t.Errorf("server error: %v", err)
			}
		}()

		// Wait for server to start and get port
		var port int
		select {
		case port = <-portCh:
			// got port
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for server to start")
		}

		// Verify port is the first available one
		if port != firstFreePort {
			t.Errorf("expected port %d (first free port), got %d", firstFreePort, port)
		}

		server.Stop()
	})

	t.Run("FallbackPort", func(t *testing.T) {
		// Find two consecutive free ports
		var firstPort, secondPort int
		for p := defaultPort; p < defaultPort+9; p++ {
			if l1, err1 := net.Listen("tcp", fmt.Sprintf(":%d", p)); err1 == nil {
				if l2, err2 := net.Listen("tcp", fmt.Sprintf(":%d", p+1)); err2 == nil {
					l1.Close()
					l2.Close()
					firstPort = p
					secondPort = p + 1
					break
				} else {
					l1.Close()
				}
			}
		}
		if firstPort == 0 {
			t.Skip("no consecutive free ports available")
		}

		// First occupy the first port
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", firstPort))
		if err != nil {
			t.Fatalf("failed to occupy port %d: %v", firstPort, err)
		}
		defer l.Close()

		// Create server
		server := NewServer(nil)

		// Channel to receive the port
		portCh := make(chan int, 1)
		server.SetPortNotifier(portCh)

		// Start server in background
		go func() {
			if err := server.Serve(tree); err != nil && err != http.ErrServerClosed {
				t.Errorf("server error: %v", err)
			}
		}()

		// Wait for server to start and get port
		var port int
		select {
		case port = <-portCh:
			// got port
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for server to start")
		}

		// Verify port is the next available one
		if port != secondPort {
			t.Errorf("expected port %d, got %d", secondPort, port)
		}

		server.Stop()
	})

	// Create server for main tests
	server := NewServer(nil) // use default renderer

	// Channel to receive the port
	portCh := make(chan int, 1)
	server.SetPortNotifier(portCh)

	// Start server in background
	go func() {
		if err := server.Serve(tree); err != nil && err != http.ErrServerClosed {
			t.Errorf("server error: %v", err)
		}
	}()

	// Wait for server to start and get port
	var port int
	select {
	case port = <-portCh:
		// got port
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for server to start")
	}

	// Helper function to make request
	makeRequest := func() (*http.Response, error) {
		return http.Get(fmt.Sprintf("http://localhost:%d", port))
	}

	t.Run("ServesSVG", func(t *testing.T) {
		// Make request to server
		resp, err := makeRequest()
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check content type
		contentType := resp.Header.Get("Content-Type")
		if contentType != "image/svg+xml" {
			t.Errorf("expected Content-Type 'image/svg+xml', got %q", contentType)
		}

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response: %v", err)
		}

		svg := string(body)

		// Basic SVG validation
		if !strings.HasPrefix(svg, "<svg") {
			t.Error("response does not start with <svg")
		}
		if !strings.HasSuffix(svg, "</svg>") {
			t.Error("response does not end with </svg>")
		}

		// Check if nodes are present
		if !strings.Contains(svg, "Root Node") {
			t.Error("SVG does not contain root node label")
		}
		if !strings.Contains(svg, "Child 1") {
			t.Error("SVG does not contain child node label")
		}
	})

	t.Run("UpdateTree", func(t *testing.T) {
		// Update tree
		newTree := &decision_tree.Node{
			ID:    "new_root",
			Label: "New Root",
		}
		server.UpdateTree(newTree)

		// Make request to server
		resp, err := makeRequest()
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response: %v", err)
		}

		svg := string(body)

		// Check if new node is present
		if !strings.Contains(svg, "New Root") {
			t.Error("SVG does not contain updated node label")
		}
	})

	// Stop server
	if err := server.Stop(); err != nil {
		t.Errorf("failed to stop server: %v", err)
	}
}

func ExampleServer_Serve() {
	// Create a simple tree
	tree := &decision_tree.Node{
		ID:    "root",
		Label: "Root",
		Children: []*decision_tree.Node{
			{
				ID:    "child",
				Label: "Child",
			},
		},
	}

	// Create and start server
	server := NewServer(nil)
	go func() {
		if err := server.Serve(tree); err != http.ErrServerClosed {
			fmt.Printf("server error: %v\n", err)
		}
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Stop server after demonstration
	server.Stop()

	// Output will be something like:
	// SVG server running at: http://localhost:12137
	// Server stop requested...
}
