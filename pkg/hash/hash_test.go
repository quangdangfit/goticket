package hash

import "testing"

func TestPasswordRoundTrip(t *testing.T) {
	h, err := Password("password1")
	if err != nil {
		t.Fatal(err)
	}
	if err := Compare(h, "password1"); err != nil {
		t.Fatalf("compare: %v", err)
	}
	if err := Compare(h, "wrong"); err == nil {
		t.Fatal("want mismatch error")
	}
}
