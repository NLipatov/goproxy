package Cookie

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestCookieBuilder_BuildCookie(t *testing.T) {
	_ = os.Setenv("ENVIRONMENT", "production")
	defer func() {
		_ = os.Unsetenv("ENVIRONMENT")
	}()

	builder := NewCookieBuilder()

	path := "/test-path"
	name := "test-cookie"
	value := "test-value"
	ttl := 10 * time.Minute

	cookie := builder.BuildCookie(path, name, value, ttl)

	if cookie.Name != name {
		t.Errorf("expected cookie name %s, got %s", name, cookie.Name)
	}
	if cookie.Value != value {
		t.Errorf("expected cookie value %s, got %s", value, cookie.Value)
	}
	if cookie.Path != path {
		t.Errorf("expected cookie path %s, got %s", path, cookie.Path)
	}
	if !cookie.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if !cookie.Secure {
		t.Error("expected Secure to be true in production")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("expected SameSiteStrictMode, got %v", cookie.SameSite)
	}
}

func TestCookieBuilder_BuildCookie_Development(t *testing.T) {
	_ = os.Setenv("ENVIRONMENT", "development")
	defer func() {
		_ = os.Unsetenv("ENVIRONMENT")
	}()

	builder := NewCookieBuilder()

	path := "/test-path"
	name := "test-cookie"
	value := "test-value"
	ttl := 10 * time.Minute

	cookie := builder.BuildCookie(path, name, value, ttl)

	if !cookie.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if cookie.Secure {
		t.Error("expected Secure to be false in development")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSiteLaxMode, got %v", cookie.SameSite)
	}
}

func TestCookieBuilder_cookieSecureFlag(t *testing.T) {
	tests := []struct {
		env          string
		expectedFlag bool
	}{
		{"production", true},
		{"development", false},
		{"staging", true},
	}

	defer func() {
		_ = os.Unsetenv("ENVIRONMENT")
	}()

	for _, tt := range tests {
		_ = os.Setenv("ENVIRONMENT", tt.env)

		builder := NewCookieBuilder()
		if builder.cookieSecureFlag() != tt.expectedFlag {
			t.Errorf("for env %s, expected secure flag %v, got %v",
				tt.env, tt.expectedFlag, builder.cookieSecureFlag())
		}
	}
}

func TestCookieBuilder_cookieSameSite(t *testing.T) {
	tests := []struct {
		env              string
		expectedSameSite http.SameSite
	}{
		{"production", http.SameSiteStrictMode},
		{"development", http.SameSiteLaxMode},
		{"staging", http.SameSiteStrictMode},
	}

	defer func() {
		_ = os.Unsetenv("ENVIRONMENT")
	}()
	for _, tt := range tests {
		_ = os.Setenv("ENVIRONMENT", tt.env)
		builder := NewCookieBuilder()
		if builder.cookieSameSite() != tt.expectedSameSite {
			t.Errorf("for env %s, expected SameSite %v, got %v",
				tt.env, tt.expectedSameSite, builder.cookieSameSite())
		}
	}
}
