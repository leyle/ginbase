package util

import "testing"

func TestGetCurTime(t *testing.T) {
	cur := GetCurTime()
	t.Log(cur.Millisecond, cur.HumanTime)
}
