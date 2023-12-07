package main

import "testing"

func Test_createDB(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "standard"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := CreateDB("./testdata/meteora.db")
			db.Close()
		})
	}
}
