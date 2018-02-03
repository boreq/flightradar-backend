package utils

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	if Slugify("abc-dEf-123") != "abc-def-123" {
		t.FailNow()
	}

	if Slugify("-abc-dEf-123-") != "abc-def-123" {
		t.FailNow()
	}

	if Slugify("-aA@$bc-d$ef-1ł2ó3-") != "aa-bc-d-ef-1-2-3" {
		t.FailNow()
	}

}
