package server

import "testing"

func TestRouteHasMissingHop(t *testing.T) {
	for _, tc := range []struct {
		hops []string
		want bool
	}{
		{nil, true},
		{[]string{"192.0.2.1", ""}, true},
		{[]string{"192.0.2.1", "192.0.2.2"}, false},
	} {
		if got := routeHasMissingHop(tc.hops); got != tc.want {
			t.Fatalf("routeHasMissingHop(%v) = %v, want %v", tc.hops, got, tc.want)
		}
	}
}

func TestBestRouteHopsNeverSplicesPaths(t *testing.T) {
	primary := []string{"192.0.2.1", "", "192.0.2.3"}
	retry := []string{"198.51.100.1", "198.51.100.2", "198.51.100.3"}
	got := bestRouteHops(primary, retry)
	if got[0] != retry[0] || got[1] != retry[1] || got[2] != retry[2] {
		t.Fatalf("best path was not a single observed pass: %v", got)
	}
	if got := bestRouteHops(retry, primary); got[0] != retry[0] {
		t.Fatalf("less complete retry displaced the first path: %v", got)
	}
}
