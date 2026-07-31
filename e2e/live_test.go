package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const baseURL = "http://127.0.0.1:3000"

// TestLiveSiteReachable is a true end-to-end check against the running
// FastToken deployment: the marketing/login page must render with the brand.
func TestLiveSiteReachable(t *testing.T) {
	resp, err := http.Get(baseURL + "/")
	if err != nil {
		t.Skip("live server unreachable: " + err.Error())
	}
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	require.Contains(t, string(body), "FastToken")
}

// TestLivePaymentStatus hits the public payment-status endpoint of the live
// gateway and asserts it responds with a structured health payload.
func TestLivePaymentStatus(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/payment/status")
	if err != nil {
		t.Skip("live server unreachable: " + err.Error())
	}
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var data map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&data))
	_, ok := data["ready"]
	require.True(t, ok, "payment status payload should include 'ready'")
}
