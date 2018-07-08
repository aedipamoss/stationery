package page

import "testing"

func TestTimestamp(t *testing.T) {
	page := Page{}
	stamp := "2018-03-22"
	expected := "[@ 2018-03-22](#2018-03-22)"
	if expected != page.Timestamp(stamp) {
		t.Errorf("expected %v, got %v", expected, page.Timestamp(stamp))
	}
}
