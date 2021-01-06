package main

import "testing"

func TestListRepository(t *testing.T) {
	if got := ListRepository(); len(got) == 0 {
		t.Errorf("No repository found")
	} else {
		t.Logf("Found repositories: %s", got)
	}
}
