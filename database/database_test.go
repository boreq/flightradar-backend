package database

import (
	"testing"
)

func TestMapperFunc(t *testing.T) {
	if mapperFunc("columnName") != "column_name" {
		t.FailNow()
	}
	if mapperFunc("columnname") != "columnname" {
		t.FailNow()
	}
	if mapperFunc("Columnname") != "columnname" {
		t.FailNow()
	}
	if mapperFunc("columnnamE") != "columnname" {
		t.FailNow()
	}
}
