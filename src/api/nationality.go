package main

// FifaCountryCodes maps 3-letter FIFA country codes to full country names.
var FifaCountryCodes = map[string]string{
	"ENG": "England", "GER": "Germany", "ESP": "Spain", "ITA": "Italy", "FRA": "France",
	"NED": "Netherlands", "POR": "Portugal", "BEL": "Belgium", "ARG": "Argentina", "BRA": "Brazil",
	"URU": "Uruguay", "COL": "Colombia", "CHI": "Chile", "MEX": "Mexico", "USA": "United States",
	"CAN": "Canada", "JPN": "Japan", "KOR": "South Korea", "AUS": "Australia", "CRO": "Croatia",
	"SUI": "Switzerland", "SWE": "Sweden", "NOR": "Norway", "DEN": "Denmark", "POL": "Poland",
	"AUT": "Austria", "TUR": "Turkey", "RUS": "Russia", "UKR": "Ukraine", "SRB": "Serbia",
	"SCO": "Scotland", "WAL": "Wales", "NIR": "Northern Ireland", "IRL": "Republic of Ireland",
	"CZE": "Czech Republic", "SVK": "Slovakia", "HUN": "Hungary", "ROU": "Romania", "GRE": "Greece",
	"EGY": "Egypt", "NGA": "Nigeria", "SEN": "Senegal", "CIV": "Ivory Coast", "GHA": "Ghana",
	"CMR": "Cameroon", "MAR": "Morocco", "ALG": "Algeria", "TUN": "Tunisia",
	// Add more as needed
}

// FifaToISO2 maps 3-letter FIFA codes to ISO 3166-1 alpha-2 country codes (or similar common web codes).
// Note: Some FIFA codes don't have direct ISO equivalents (e.g., ENG, SCO, WAL, NIR are part of GB in ISO).
// Using common web representations like 'gb-eng'.
var FifaToISO2 = map[string]string{
	"ENG": "gb-eng", "GER": "de", "ESP": "es", "ITA": "it", "FRA": "fr", "NED": "nl",
	"POR": "pt", "BEL": "be", "ARG": "ar", "BRA": "br", "URU": "uy", "COL": "co",
	"CHI": "cl", "MEX": "mx", "USA": "us", "CAN": "ca", "JPN": "jp", "KOR": "kr",
	"AUS": "au", "CRO": "hr", "SUI": "ch", "SWE": "se", "NOR": "no", "DEN": "dk",
	"POL": "pl", "AUT": "at", "TUR": "tr", "RUS": "ru", "UKR": "ua", "SRB": "rs",
	"SCO": "gb-sct", "WAL": "gb-wls", "NIR": "gb-nir", "IRL": "ie", "CZE": "cz",
	"SVK": "sk", "HUN": "hu", "ROU": "ro", "GRE": "gr", "EGY": "eg", "NGA": "ng",
	"SEN": "sn", "CIV": "ci", "GHA": "gh", "CMR": "cm", "MAR": "ma", "ALG": "dz",
	"TUN": "tn",
	// Add more as needed
}
