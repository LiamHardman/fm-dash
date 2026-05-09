package main

import (
	"strings"
	"testing"
)

func TestResolveClubLogoUsesAliasWithExternalImageAPI(t *testing.T) {
	t.Setenv("IMAGE_API_URL", "https://cdn.example/uploads")

	resolution := resolveClubLogo("Manchester United")

	if resolution.Status != clubLogoStatusExact {
		t.Fatalf("expected exact status, got %q (%s)", resolution.Status, resolution.Reason)
	}
	if resolution.TeamID != "680" {
		t.Fatalf("expected Man Utd team ID 680, got %q", resolution.TeamID)
	}
	if !strings.Contains(resolution.LogoURL, "teamId=680") {
		t.Fatalf("expected logo URL to reference team ID 680, got %q", resolution.LogoURL)
	}
}

func TestResolveClubLogoReturnsUnmatchedForUnknownClub(t *testing.T) {
	t.Setenv("IMAGE_API_URL", "https://cdn.example/uploads")

	resolution := resolveClubLogo("Not A Real Club Name 12345")

	if resolution.Status != clubLogoStatusUnmatched {
		t.Fatalf("expected unmatched status, got %q", resolution.Status)
	}
	if resolution.LogoURL != "" {
		t.Fatalf("expected no logo URL for unmatched club, got %q", resolution.LogoURL)
	}
}

func TestResolveClubLogoFindsLowercaseLocalLogoPath(t *testing.T) {
	t.Setenv("IMAGE_API_URL", "")
	t.Setenv("LOGOS_DIR", "../../logos")

	resolution := resolveClubLogo("Arsenal")

	if resolution.Status != clubLogoStatusExact && resolution.Status != clubLogoStatusConfident {
		t.Fatalf("expected displayable status, got %q (%s)", resolution.Status, resolution.Reason)
	}
	if resolution.TeamID != "602" {
		t.Fatalf("expected Arsenal team ID 602, got %q", resolution.TeamID)
	}
	if resolution.LogoURL == "" {
		t.Fatal("expected local Arsenal logo to be displayable")
	}
}
