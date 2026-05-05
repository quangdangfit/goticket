package jwt

import (
	"testing"
	"time"
)

func TestManager_IssueAndVerify(t *testing.T) {
	m := New("access-secret", "refresh-secret", time.Minute, time.Hour)
	tok, _, err := m.IssueAccess("user-1", "admin")
	if err != nil {
		t.Fatal(err)
	}
	uid, role, err := m.Verify(tok)
	if err != nil {
		t.Fatal(err)
	}
	if uid != "user-1" || role != "admin" {
		t.Fatalf("got %s/%s", uid, role)
	}
}

func TestManager_VerifyExpired(t *testing.T) {
	m := New("access-secret", "refresh-secret", -time.Second, time.Hour)
	tok, _, _ := m.IssueAccess("user-1", "user")
	if _, _, err := m.Verify(tok); err == nil {
		t.Fatal("want error for expired token")
	}
}

func TestManager_HashRefreshIsStable(t *testing.T) {
	m := New("a", "b", time.Minute, time.Hour)
	h1 := m.HashRefresh("xyz")
	h2 := m.HashRefresh("xyz")
	if h1 != h2 || h1 == "" {
		t.Fatalf("unstable: %q vs %q", h1, h2)
	}
}
