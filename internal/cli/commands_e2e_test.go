package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIAddCatLs performs an end-to-end test of the add, cat, and ls CLI commands.
func TestCLIAddCatLs(t *testing.T) {
	// Set up a temporary working directory.
	tmpDir, err := os.MkdirTemp("", "p2pfs_e2e")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Switch to temp directory for datastore and CLI operations.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create an input file.
	inputFile := filepath.Join(tmpDir, "input.txt")
	if err := os.WriteFile(inputFile, []byte("hello e2e"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run `p2pfs add`
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"add", inputFile})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("add failed: %v, output: %s", err, buf.String())
	}
	cid := strings.TrimSpace(buf.String())
	if cid == "" {
		t.Fatalf("expected CID from add, got empty output")
	}
	buf.Reset()

	// Run `p2pfs cat`
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"cat", cid})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("cat failed: %v", err)
	}
	if buf.String() != "hello e2e" {
		t.Fatalf("expected 'hello e2e', got '%s'", buf.String())
	}
	buf.Reset()

	// Run `p2pfs ls` (should produce no links for a raw block)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"ls", cid})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("ls failed: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "" {
		t.Fatalf("expected empty ls output, got '%s'", buf.String())
	}
}
