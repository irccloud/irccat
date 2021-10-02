package util

import (
	"os"
	"path"
	"testing"
	"time"
)

func TestIsExpired(t *testing.T) {
	ts := time.Now().Unix()
	if isExpired(ts, time.Second) {
		t.Error("Shouldn't have expired")
	}
	if !isExpired(ts, time.Microsecond) {
		t.Error("Should've expired")
	}
}

func TestCheckTimestamp(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	file := path.Join(dir, "timestamp")
	WriteTimestamp(file, now)
	if err := CheckTimestamp(file, time.Second); err != nil {
		t.Error(err)
	}
	err := CheckTimestamp(file, time.Microsecond)
	if err == nil {
		t.Error("Did not raise timeout")
	}
	t.Log(err)
	fake := path.Join(dir, "fake")
	// File missing
	err = CheckTimestamp(fake, time.Second)
	if err == nil {
		t.Error("Did not raise missing file error")
	}
	t.Log(err)
	// Wrong format
	if err := os.WriteFile(fake, []byte{':', '('}, 0600); err != nil {
		t.Error(err)
	}
	err = CheckTimestamp(fake, time.Second)
	if err == nil {
		t.Error("Did not raise bad parse error")
	}
	t.Log(err)
}
