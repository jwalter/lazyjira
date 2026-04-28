package auth

import "testing"

func TestBasicAuthHeader(t *testing.T) {
	t.Parallel()

	got := BasicAuthHeader("user@example.com", "secret-token")
	want := "Basic dXNlckBleGFtcGxlLmNvbTpzZWNyZXQtdG9rZW4="

	if got != want {
		t.Fatalf("BasicAuthHeader() = %q, want %q", got, want)
	}
}
