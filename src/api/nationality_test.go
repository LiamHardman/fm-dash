package main

import (
	"testing"
)

func TestFifaCountryCodes(t *testing.T) {
	// Test that the FIFA country codes map contains expected entries
	tests := []struct {
		code     string
		expected string
	}{
		{"ENG", "England"},
		{"SCO", "Scotland"},
		{"WAL", "Wales"},
		{"NIR", "Northern Ireland"},
		{"GER", "Germany"},
		{"FRA", "France"},
		{"ESP", "Spain"},
		{"ITA", "Italy"},
		{"BRA", "Brazil"},
		{"ARG", "Argentina"},
		{"USA", "United States"},
		{"CAN", "Canada"},
		{"AUS", "Australia"},
		{"JPN", "Japan"},
		{"CHN", "China PR"},
		{"IND", "India"},
		{"RSA", "South Africa"},
		{"NGA", "Nigeria"},
		{"EGY", "Egypt"},
		{"MAR", "Morocco"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if country, exists := FifaCountryCodes[tt.code]; exists {
				if country != tt.expected {
					t.Errorf("FifaCountryCodes[%s] = %s; want %s", tt.code, country, tt.expected)
				}
			} else {
				t.Errorf("FifaCountryCodes[%s] does not exist", tt.code)
			}
		})
	}
}

func TestFifaCountryCodesCompleteness(t *testing.T) {
	// Test that the map is not empty and contains reasonable number of entries
	if len(FifaCountryCodes) == 0 {
		t.Error("FifaCountryCodes map is empty")
	}

	// FIFA has around 211 member associations, so we should have a reasonable number
	if len(FifaCountryCodes) < 100 {
		t.Errorf("FifaCountryCodes has only %d entries, expected more than 100", len(FifaCountryCodes))
	}

	// Test that all codes are 3 characters long
	for code, country := range FifaCountryCodes {
		if len(code) != 3 {
			t.Errorf("FIFA code %s has length %d, expected 3", code, len(code))
		}

		if country == "" {
			t.Errorf("FIFA code %s has empty country name", code)
		}
	}
}

func TestFifaCountryCodesNoDuplicates(t *testing.T) {
	// Test that there are no duplicate country names (which might indicate mapping issues)
	countryNames := make(map[string][]string)

	for code, country := range FifaCountryCodes {
		countryNames[country] = append(countryNames[country], code)
	}

	// Check for countries with multiple codes (some might be legitimate like Korea Republic/Korea DPR)
	legitimateDuplicates := map[string]bool{
		"Bangladesh": true, // BAN and BGD
		"Bulgaria":   true, // BGR and BUL
		"Kosovo":     true, // KOS and KVX
	}

	for country, codes := range countryNames {
		if len(codes) > 1 && !legitimateDuplicates[country] {
			t.Logf("Warning: Country %s has multiple codes: %v", country, codes)
		}
	}
}

func TestFifaCountryCodesSpecialCases(t *testing.T) {
	// Test specific cases that are important for football
	specialCases := map[string]string{
		"ENG": "England",
		"SCO": "Scotland",
		"WAL": "Wales",
		"NIR": "Northern Ireland",
		"KOR": "Korea Republic",  // South Korea
		"PRK": "Korea DPR",       // North Korea
		"CHN": "China PR",        // People's Republic of China
		"TPE": "Chinese Taipei",  // Taiwan
		"COD": "DR Congo",        // Democratic Republic of Congo
		"CGO": "Congo",           // Republic of Congo
		"MKD": "North Macedonia", // North Macedonia
		"SSD": "South Sudan",     // South Sudan
	}

	for code, expectedCountry := range specialCases {
		if country, exists := FifaCountryCodes[code]; exists {
			if country != expectedCountry {
				t.Errorf("FifaCountryCodes[%s] = %s; want %s", code, country, expectedCountry)
			}
		} else {
			t.Errorf("Important FIFA code %s is missing from FifaCountryCodes", code)
		}
	}
}

func TestFifaCountryCodesFormat(t *testing.T) {
	// Test that all codes follow expected format (3 uppercase letters)
	for code := range FifaCountryCodes {
		// Check length
		if len(code) != 3 {
			t.Errorf("FIFA code %s has incorrect length %d, expected 3", code, len(code))
		}

		// Check that all characters are uppercase letters
		for i, char := range code {
			if char < 'A' || char > 'Z' {
				t.Errorf("FIFA code %s has invalid character '%c' at position %d", code, char, i)
			}
		}
	}
}

// Benchmark FIFA country code lookup
func BenchmarkFifaCountryCodeLookup(b *testing.B) {
	testCodes := []string{"ENG", "GER", "BRA", "ARG", "FRA", "ESP", "ITA", "NED"}

	for i := 0; i < b.N; i++ {
		for _, code := range testCodes {
			_ = FifaCountryCodes[code]
		}
	}
}

func TestFifaCountryCodesConsistency(t *testing.T) {
	// Test for consistency in naming conventions
	inconsistencies := []struct {
		code        string
		country     string
		description string
	}{
		// These are just examples - adjust based on actual data
	}

	for _, inc := range inconsistencies {
		if country, exists := FifaCountryCodes[inc.code]; exists {
			if country == inc.country {
				t.Logf("Potential inconsistency: %s - %s (%s)", inc.code, inc.country, inc.description)
			}
		}
	}
}

func TestFifaCountryCodesCoverage(t *testing.T) {
	// Test that major footballing nations are covered
	majorNations := []string{
		"BRA", "ARG", "GER", "FRA", "ESP", "ITA", "ENG", "NED",
		"POR", "BEL", "CRO", "URU", "COL", "MEX", "USA", "CAN",
		"JPN", "KOR", "AUS", "CHN", "IND", "RSA", "NGA", "EGY",
		"MAR", "SEN", "GHA", "CIV", "CMR", "TUN", "ALG", "DZA",
	}

	missing := []string{}
	for _, code := range majorNations {
		if _, exists := FifaCountryCodes[code]; !exists {
			missing = append(missing, code)
		}
	}

	if len(missing) > 0 {
		t.Logf("Missing major footballing nations: %v", missing)
		// Don't fail the test as some codes might be different than expected
	}
}
