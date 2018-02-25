package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestStationery(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestStationery")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	} else {
		t.Fatalf("process ran with err %v, want exit status 1", err)
	}
}
