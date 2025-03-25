package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsHTTPText(t *testing.T) {
	header := http.Header{}
	header.Set("Content-Type", "text/plain")
	if !IsHTTPText(header) {
		t.Errorf("IsHTTPText(%v) = false; want true", header)
	}

	header.Set("Content-Type", "application/octet-stream")
	if IsHTTPText(header) {
		t.Errorf("IsHTTPText(%v) = true; want false", header)
	}
}

func TestBodyToString(t *testing.T) {
	header := http.Header{}
	header.Set("Content-Type", "text/plain")
	body := []byte("test")
	result := BodyToString(header, body)
	if result != "test" {
		t.Errorf("BodyToString(%v, %v) = %s; want test", header, body, result)
	}

	header.Set("Content-Type", "application/octet-stream")
	result = BodyToString(header, body)
	expected := fmt.Sprintf("<Binary Data: %d bytes>", len(body))
	if result != expected {
		t.Errorf("BodyToString(%v, %v) = %s; want %s", header, body, result, expected)
	}
}

func TestGetAbsolutePath(t *testing.T) {
	p := "./test"
	base := "/home/user"
	result := GetAbsolutePath(p, base)
	expected := filepath.Clean(filepath.Join(base, p))
	if result != expected {
		t.Errorf("GetAbsolutePath(%s, %s) = %s; want %s", p, base, result, expected)
	}

	p = "/absolute/path"
	result = GetAbsolutePath(p, base)
	expected = filepath.Clean(p)
	if result != expected {
		t.Errorf("GetAbsolutePath(%s, %s) = %s; want %s", p, base, result, expected)
	}
}

func TestGetUserDataDir(t *testing.T) {
	dir, err := GetUserDataDir()
	if err != nil {
		t.Errorf("GetUserDataDir() error = %v", err)
	}
	var expected string
	switch sys := runtime.GOOS; sys {
	case "darwin":
		expected = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	case "windows":
		expected = filepath.Join(os.Getenv("APPDATA"))
	case "linux":
		expected = filepath.Join(os.Getenv("HOME"), ".config")
	default:
		t.Errorf("unsupported operating system")
	}
	if dir != expected {
		t.Errorf("GetUserDataDir() = %s; want %s", dir, expected)
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	target := "b"
	if !Contains(slice, target) {
		t.Errorf("Contains(%v, %s) = false; want true", slice, target)
	}

	target = "d"
	if Contains(slice, target) {
		t.Errorf("Contains(%v, %s) = true; want false", slice, target)
	}
}

func TestSha256hex(t *testing.T) {
	s := "test"
	result := Sha256hex(s)
	// Convert [32]byte to []byte
	hash := sha256.Sum256([]byte(s))
	expected := hex.EncodeToString(hash[:])
	if result != expected {
		t.Errorf("Sha256hex(%s) = %s; want %s", s, result, expected)
	}
}

func TestHmacSha256(t *testing.T) {
	s := "test"
	key := "key"
	result := HmacSha256(s, key)
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(s))
	expected := string(h.Sum(nil))
	if result != expected {
		t.Errorf("HmacSha256(%s, %s) = %s; want %s", s, key, result, expected)
	}
}
