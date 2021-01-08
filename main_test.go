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

//func TestGetFileSize(t *testing.T) {
//	size, err := gddl.GetFileSize("nlp", "決算短信", "2020-sorted")
//	if err != nil {
//		t.Errorf("Error: %v", err)
//	}
//	t.Log(size)
//}
