package util

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func isExpired(lastStamp int64, period time.Duration) bool {
	return time.Now().Sub(time.Unix(lastStamp, 0)) > period
}

func getTimestamp(tsFile string) (int64, error) {
	raw, err := os.ReadFile(tsFile)
	if err != nil {
		return 0, fmt.Errorf("Couldn't read timestamp file: %s", err)
	}
	s := strings.TrimRight(string(raw), "\n")
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// Parsing error already includes offending text
		return 0, fmt.Errorf("Couldn't parse timestamp: %s", err)
	}
	return ts, nil
}

// WriteTimestamp creates a timestamp file.
func WriteTimestamp(tsFile string, t time.Time) error {
	s := fmt.Sprintf("%d\n", t.Unix())
	err := os.WriteFile(tsFile, []byte(s), 0666)
	if err != nil {
		return fmt.Errorf("Couldn't write to timestamp file: %s", err)
	}
	return nil
}

// CheckTimestamp reads a timestamp file and returns an error if it's expired.
//
// The file should contain a decimal representation of a Unix timestamp and an
// optional LF newline.
func CheckTimestamp(tsFile string, period time.Duration) error {
	ts, err := getTimestamp(tsFile)
	if err != nil {
		return err
	}
	if isExpired(ts, period) {
		diff := time.Now().Sub(time.Unix(ts, 0).Add(period))
		return fmt.Errorf("Timeout exceeded by %s", diff.String())
	}
	return nil
}
