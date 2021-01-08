package main

import (
	"github.com/masanorihirano/gddl/gddl"
	"testing"
)

func TestListRepository(t *testing.T) {
	if got := gddl.ListRepository(); len(got) == 0 {
		t.Errorf("No repository found")
	} else {
		t.Logf("Found repositories: %s", got)
	}
}
