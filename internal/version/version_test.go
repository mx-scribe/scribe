package version

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestInfo(t *testing.T) {
	info := Info()
	if info == "" {
		t.Error("Info() should not return empty string")
	}
	if info != Version {
		t.Errorf("Info() = %q, want %q", info, Version)
	}
}

func TestFull(t *testing.T) {
	full := Full()
	if full == "" {
		t.Error("Full() should not return empty string")
	}
	if !strings.Contains(full, Version) {
		t.Errorf("Full() should contain version, got %q", full)
	}
	if !strings.Contains(full, "commit:") {
		t.Errorf("Full() should contain commit info, got %q", full)
	}
	if !strings.Contains(full, "built:") {
		t.Errorf("Full() should contain build info, got %q", full)
	}
}
