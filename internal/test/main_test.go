// Package test provides integration tests for caos server implementations.
package test

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

var serverPort int
const testAPIKey = "test-api-key-123"

// TestMain manages the server lifecycle for integration tests.
func TestMain(m *testing.M) {
	port := pickFreePort()
	serverPort = port
	root := "/tmp/caos-test-" + strconv.Itoa(os.Getpid())

	// Build the server binary
	repoDir := "caos.one/caos/cmd/caos"
	cmd := exec.Command("go", "build", "-o", "/tmp/caos-test-bin", repoDir)
	cmd.Dir = "/home/dknr/src/caos"
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build server: %v\n", err)
		os.Exit(1)
	}

	// Start the server with API key protection
	serverCmd := exec.Command("/tmp/caos-test-bin", "serve",
		"--port", strconv.Itoa(port),
		"--root", root,
		"--api-key", testAPIKey)
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	if err := serverCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

	// Wait for server to be ready
	base := fmt.Sprintf("http://localhost:%d", port)
	hc := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	for i := 0; i < 30; i++ {
		resp, err := hc.Get(base + "/")
		if err == nil && resp.StatusCode == 302 {
			resp.Body.Close()
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Run tests
	exitCode := m.Run()

	// Cleanup
	serverCmd.Process.Signal(os.Interrupt)
	serverCmd.Wait()
	os.RemoveAll(root)

	os.Exit(exitCode)
}

// pickFreePort finds an available TCP port.
func pickFreePort() int {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 31923
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

// TestCaosSuite runs the full integration test suite.
func TestCaosSuite(t *testing.T) {
	base := fmt.Sprintf("http://localhost:%d", serverPort)
	c := NewSuiteClient(base, testAPIKey)

	RunSuite(t, c)
}
