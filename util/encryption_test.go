package util

import (
	"encoding/base64"
	"testing"
)

func TestEncrypt(t *testing.T) {
	bin, _ := base64.RawStdEncoding.DecodeString("mD0I8u2luw52GIQpEteYWQu2UxsWP4kacSBhjgAh5C9")
	key = string(bin)
	keyPresent = true

	out, err := Encrypt("sources-api-tests")
	if err != nil {
		t.Error(err)
	}

	if out != "v2:{f80zhczL8GTPqCdlU8WX7m+8BgCZgwXERNGYUF7J+lU}" {
		t.Errorf("encryption failed, got %v expected %v", out, "v2:{f80zhczL8GTPqCdlU8WX7m+8BgCZgwXERNGYUF7J+lU}")
	}
}

func TestDecrypt(t *testing.T) {
	bin, _ := base64.RawStdEncoding.DecodeString("mD0I8u2luw52GIQpEteYWQu2UxsWP4kacSBhjgAh5C9")
	key = string(bin)
	keyPresent = true

	out, err := Decrypt("v2:{f80zhczL8GTPqCdlU8WX7m+8BgCZgwXERNGYUF7J+lU}")
	if err != nil {
		t.Error(err)
	}

	if out != "sources-api-tests" {
		t.Errorf("decryption failed, got %v expected %v", out, "sources-api-tests")
	}
}

func TestBadFormat(t *testing.T) {
	bin, _ := base64.RawStdEncoding.DecodeString("mD0I8u2luw52GIQpEteYWQu2UxsWP4kacSBhjgAh5C9")
	key = string(bin)

	_, err := Decrypt("a bad thing")
	if err == nil {
		t.Errorf("'a bad thing': expected an err but none was returned")
	}

	_, err = Decrypt("v2{}")
	if err == nil {
		t.Errorf("v2{}: expected an err but none was returned")
	}
}

func TestPadding(t *testing.T) {
	out := padString("thing", 16)

	if len(out) != 16 {
		t.Errorf("padding was not added properly to string")
	}
}

func TestNoKey(t *testing.T) {
	key = ""
	keyPresent = false

	_, err := Decrypt("another thing")
	if err == nil {
		t.Errorf("expected an err but none was returned")
	}

	if err.Error() != "no encryption key present" {
		t.Errorf("bad error message: %v", err.Error())
	}
}
