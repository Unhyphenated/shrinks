//go:build unit

package util

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ===== USER-AGENT PARSING TESTS =====

// Test #9: Desktop browser detection
func TestParseUserAgent_Desktop(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	result := ParseUserAgent(ua)

	if result.DeviceType != "Desktop" {
		t.Errorf("DeviceType = %s, want Desktop", result.DeviceType)
	}
	if result.OS != "Windows" {
		t.Errorf("OS = %s, want Windows", result.OS)
	}
	if result.Browser == "" {
		t.Error("Browser should not be empty")
	}
}

// Test #10: Mobile browser detection
func TestParseUserAgent_Mobile(t *testing.T) {
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"

	result := ParseUserAgent(ua)

	if result.DeviceType != "Mobile" {
		t.Errorf("DeviceType = %s, want Mobile", result.DeviceType)
	}
	if result.OS != "iOS" {
		t.Errorf("OS = %s, want iOS", result.OS)
	}
}

// Test #11: Tablet detection
func TestParseUserAgent_Tablet(t *testing.T) {
	ua := "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"

	result := ParseUserAgent(ua)

	if result.DeviceType != "Tablet" {
		t.Errorf("DeviceType = %s, want Tablet", result.DeviceType)
	}
}

// Test #12: Bot detection
func TestParseUserAgent_Bot(t *testing.T) {
	ua := "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

	result := ParseUserAgent(ua)

	if result.DeviceType != "Bot" {
		t.Errorf("DeviceType = %s, want Bot", result.DeviceType)
	}
}

// Test #13: Empty user agent
func TestParseUserAgent_Empty(t *testing.T) {
	result := ParseUserAgent("")

	// Should not panic, should return something reasonable
	if result.DeviceType == "" {
		t.Error("DeviceType should have a fallback value")
	}
}

// ===== IP EXTRACTION TESTS =====

// Test #14: Direct connection (no proxy)
func TestGetIP_Direct(t *testing.T) {
	ip := GetIP("", "192.168.1.100:54321")

	if ip != "192.168.1.100" {
		t.Errorf("GetIP = %s, want 192.168.1.100", ip)
	}
}

// Test #15: Behind proxy (X-Forwarded-For)
func TestGetIP_Forwarded(t *testing.T) {
	ip := GetIP("203.0.113.50", "10.0.0.1:54321")

	if ip != "203.0.113.50" {
		t.Errorf("GetIP = %s, want 203.0.113.50", ip)
	}
}

// Test #16: Multiple proxies (take first)
func TestGetIP_ForwardedMultiple(t *testing.T) {
	ip := GetIP("203.0.113.50, 70.41.3.18, 150.172.238.178", "10.0.0.1:54321")

	if ip != "203.0.113.50" {
		t.Errorf("GetIP = %s, want 203.0.113.50 (first in chain)", ip)
	}
}

// ===== IP ANONYMIZATION TESTS =====

// Test #17: IPv4 anonymization (zero last octet)
func TestAnonymizeIP_IPv4(t *testing.T) {
	result := AnonymizeIP("192.168.1.100")

	if result != "192.168.1.0" {
		t.Errorf("AnonymizeIP = %s, want 192.168.1.0", result)
	}
}

// Test #18: IPv6 handling
func TestAnonymizeIP_IPv6(t *testing.T) {
	result := AnonymizeIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")

	// Should return something (either anonymized or original)
	if result == "" {
		t.Error("AnonymizeIP returned empty string for IPv6")
	}
}

// Test #19: Invalid IP returns fallback
func TestAnonymizeIP_Invalid(t *testing.T) {
	result := AnonymizeIP("not-an-ip")

	if result != "0.0.0.0" {
		t.Errorf("AnonymizeIP = %s, want 0.0.0.0 for invalid input", result)
	}
}

// ===== JSON HELPER TESTS =====

// Test #20: WriteJSON sets correct headers and body
func TestWriteJSON_Success(t *testing.T) {
	rr := httptest.NewRecorder()

	data := map[string]string{"message": "hello"}
	WriteJSON(rr, http.StatusOK, data)

	// Check status
	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}

	// Check content-type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", contentType)
	}

	// Check body is valid JSON
	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Response body is not valid JSON: %v", err)
	}

	if result["message"] != "hello" {
		t.Errorf("Body message = %s, want hello", result["message"])
	}
}

// Test #21: WriteError produces correct error format
func TestWriteError_Format(t *testing.T) {
	rr := httptest.NewRecorder()

	WriteError(rr, http.StatusBadRequest, "something went wrong")

	// Check status
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	// Check body structure
	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Response body is not valid JSON: %v", err)
	}

	if result["error"] != "something went wrong" {
		t.Errorf("Error message = %s, want 'something went wrong'", result["error"])
	}
}

// Bonus: WriteJSON with different status codes
func TestWriteJSON_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"Created", http.StatusCreated},
		{"Accepted", http.StatusAccepted},
		{"NoContent", http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			WriteJSON(rr, tt.status, map[string]string{})

			if rr.Code != tt.status {
				t.Errorf("Status = %d, want %d", rr.Code, tt.status)
			}
		})
	}
}
