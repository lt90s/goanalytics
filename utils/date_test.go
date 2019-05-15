package utils

import (
	"testing"
	"time"
)

func TestToday(t *testing.T) {
	t.Log(today)
	t.Log(today.In(location).Unix())
	t.Log(time.Now().Unix())
	t.Log(time.Now().In(location))
}
