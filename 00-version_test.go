package asock

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	vfile, err := os.Open("./RELEASE_NOTES")
	defer vfile.Close()
	if err != nil {
		t.Fatal("Couldn't open RELEASE_NOTES")
	}
	vreader := bufio.NewReader(vfile)
	vline, err := vreader.ReadString('\n')
	if err != nil {
		t.Errorf("Got error reading RELEASE_NOTES: %v", err)
	}
	chunks := strings.Split(vline, " ")
	if chunks[0] != Version {
		t.Errorf("RELEASE_NOTES says version %v but package Version is %v", chunks[0], Version)
	}
}
