package database

import "testing"

func Test_createDB(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "standard"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := createDB("../testdata/meteora.db")
			db.Close()
		})
	}
}