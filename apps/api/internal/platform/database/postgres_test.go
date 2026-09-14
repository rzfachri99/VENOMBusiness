package database

import "testing"

func TestRebind(t *testing.T) {
	got := rebind("SELECT * FROM x WHERE a=? AND b=? AND note='?' ")
	want := "SELECT * FROM x WHERE a=$1 AND b=$2 AND note='?' "
	if got != want {
		t.Fatalf("rebind: got %q want %q", got, want)
	}
}
