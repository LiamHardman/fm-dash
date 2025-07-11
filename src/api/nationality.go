package main

// FifaCountryCodes maps 3-letter FIFA country codes to full country names.
var FifaCountryCodes = map[string]string{
	"AFG": "Afghanistan",
	"AIA": "Anguilla",
	"ALB": "Albania",
	"ALG": "Algeria",
	"AND": "Andorra",
	"ANG": "Angola",
	"ANT": "Netherlands Antilles", // Historical, consider if still needed or use alternatives like CUW for Curaçao if applicable
	"ARG": "Argentina",
	"ARM": "Armenia",
	"ARU": "Aruba",
	"ASA": "American Samoa",
	"ATG": "Antigua and Barbuda",
	"AUS": "Australia",
	"AUT": "Austria",
	"AZE": "Azerbaijan",
	"BAH": "Bahamas",
	"BAN": "Bangladesh",
	"BDI": "Burundi",
	"BEL": "Belgium",
	"BEN": "Benin",
	"BER": "Bermuda",
	"BFA": "Burkina Faso",
	"BGD": "Bangladesh", // Note: BAN is also listed for Bangladesh, choose one for consistency
	"BGR": "Bulgaria",
	"BHU": "Bhutan",
	"BIH": "Bosnia and Herzegovina",
	"BLR": "Belarus",
	"BLZ": "Belize",
	"BOL": "Bolivia",
	"BOT": "Botswana",
	"BRA": "Brazil",
	"BRB": "Barbados",
	"BRU": "Brunei Darussalam",
	"BUL": "Bulgaria", // Already listed as BGR, choose one
	"BHR": "Bahrain",
	"CAM": "Cambodia",
	"CAN": "Canada",
	"CAY": "Cayman Islands",
	"CGO": "Congo",
	"CHA": "Chad",
	"CHI": "Chile",
	"CHN": "China PR",
	"CIV": "Côte d'Ivoire", // Ivory Coast
	"CMR": "Cameroon",
	"COD": "DR Congo",
	"COK": "Cook Islands",
	"COL": "Colombia",
	"COM": "Comoros",
	"CPV": "Cape Verde",
	"CRC": "Costa Rica",
	"CRO": "Croatia",
	"CTA": "Central African Republic",
	"CUB": "Cuba",
	"CUW": "Curaçao",
	"CYP": "Cyprus",
	"CZE": "Czech Republic",
	"DEN": "Denmark",
	"DJI": "Djibouti",
	"DMA": "Dominica",
	"DOM": "Dominican Republic",
	"ECU": "Ecuador",
	"EGY": "Egypt",
	"ENG": "England",
	"EQG": "Equatorial Guinea",
	"ERI": "Eritrea",
	"ESP": "Spain",
	"EST": "Estonia",
	"ETH": "Ethiopia",
	"FIJ": "Fiji",
	"FIN": "Finland",
	"FRA": "France",
	"FRO": "Faroe Islands",
	"GAB": "Gabon",
	"GAM": "Gambia",
	"GEO": "Georgia",
	"GER": "Germany",
	"GHA": "Ghana",
	"GIB": "Gibraltar",
	"GNB": "Guinea-Bissau",
	"GRE": "Greece",
	"GRN": "Grenada",
	"GUA": "Guatemala",
	"GUI": "Guinea",
	"GUM": "Guam",
	"GUY": "Guyana",
	"HAI": "Haiti",
	"HKG": "Hong Kong",
	"HON": "Honduras",
	"HUN": "Hungary",
	"IDN": "Indonesia",
	"IND": "India",
	"IRL": "Republic of Ireland",
	"IRN": "Iran",
	"IRQ": "Iraq",
	"ISL": "Iceland",
	"ISR": "Israel",
	"ITA": "Italy",
	"JAM": "Jamaica",
	"JOR": "Jordan",
	"JPN": "Japan",
	"KAZ": "Kazakhstan",
	"KEN": "Kenya",
	"KGZ": "Kyrgyzstan",
	"KOR": "Korea Republic", // South Korea
	"KOS": "Kosovo",
	"KSA": "Saudi Arabia",
	"KUW": "Kuwait",
	"KVX": "Kosovo", // Duplicate of KOS, choose one
	"LAO": "Laos",
	"LBN": "Lebanon", // LIB is also used in one source, LBN is more common in others
	"LBR": "Liberia",
	"LBY": "Libya",
	"LCA": "Saint Lucia",
	"LES": "Lesotho",
	"LIE": "Liechtenstein",
	"LTU": "Lithuania",
	"LUX": "Luxembourg",
	"LVA": "Latvia",
	"MAC": "Macau",
	"MAD": "Madagascar",
	"MAR": "Morocco",
	"MAS": "Malaysia",
	"MDA": "Moldova",
	"MDV": "Maldives",
	"MEX": "Mexico",
	"MGL": "Mongolia", // Also listed as MNG in other sources
	"MKD": "North Macedonia",
	"MLI": "Mali",
	"MLT": "Malta",
	"MNE": "Montenegro",
	"MOZ": "Mozambique",
	"MRI": "Mauritius",
	"MSR": "Montserrat",
	"MTN": "Mauritania",
	"MWI": "Malawi",
	"MYA": "Myanmar",
	"NAM": "Namibia",
	"NCA": "Nicaragua",
	"NCL": "New Caledonia",
	"NED": "Netherlands",
	"NEP": "Nepal",
	"NGA": "Nigeria",
	"NIG": "Niger",
	"NIR": "Northern Ireland",
	"NOR": "Norway",
	"NZL": "New Zealand",
	"OMA": "Oman",
	"PAK": "Pakistan",
	"PAN": "Panama",
	"PAR": "Paraguay",
	"PER": "Peru",
	"PHI": "Philippines",
	"PLE": "Palestine",
	"PNG": "Papua New Guinea",
	"POL": "Poland",
	"POR": "Portugal",
	"PRK": "Korea DPR", // North Korea
	"PUR": "Puerto Rico",
	"QAT": "Qatar",
	"ROU": "Romania", // Also listed as ROM in one source
	"RSA": "South Africa",
	"RUS": "Russia",
	"RWA": "Rwanda",
	"SAM": "Samoa",
	"SCO": "Scotland",
	"SEN": "Senegal",
	"SEY": "Seychelles",
	"SGP": "Singapore", // Also listed as SIN in one source
	"SKN": "Saint Kitts and Nevis",
	"SLE": "Sierra Leone",
	"SLV": "El Salvador",
	"SMR": "San Marino",
	"SOL": "Solomon Islands",
	"SOM": "Somalia",
	"SRB": "Serbia",
	"SRI": "Sri Lanka",
	"SSD": "South Sudan",
	"STP": "São Tomé and Príncipe",
	"SUD": "Sudan",
	"SUI": "Switzerland",
	"SUR": "Suriname",
	"SVK": "Slovakia",
	"SVN": "Slovenia",
	"SWE": "Sweden",
	"SWZ": "Eswatini", // Swaziland
	"SYR": "Syria",
	"TAH": "Tahiti", // French Polynesia
	"TAN": "Tanzania",
	"TCA": "Turks and Caicos Islands",
	"TGA": "Tonga",
	"THA": "Thailand",
	"TJK": "Tajikistan",
	"TKM": "Turkmenistan",
	"TLS": "Timor-Leste",
	"TOG": "Togo",
	"TPE": "Chinese Taipei", // Taiwan
	"TRI": "Trinidad and Tobago",
	"TUN": "Tunisia",
	"TUR": "Turkey",
	"UAE": "United Arab Emirates",
	"UGA": "Uganda",
	"UKR": "Ukraine",
	"URU": "Uruguay",
	"USA": "United States",
	"UZB": "Uzbekistan",
	"VAN": "Vanuatu",
	"VEN": "Venezuela",
	"VGB": "British Virgin Islands",
	"VIE": "Vietnam",
	"VIN": "Saint Vincent and the Grenadines",
	"VIR": "US Virgin Islands",
	"WAL": "Wales",
	"YEM": "Yemen",
	"ZAM": "Zambia",
	"ZIM": "Zimbabwe",
	// Add any other specific codes your application might encounter or need.
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
	"AFG": "af", "ALB": "al", "AND": "ad", "ANG": "ao", "ARM": "am", "ARU": "aw",
	"ASA": "as", "ATG": "ag", "AZE": "az", "BAH": "bs", "BAN": "bd", "BDI": "bi",
	"BEN": "bj", "BER": "bm", "BFA": "bf", "BHU": "bt", "BIH": "ba", "BLR": "by",
	"BLZ": "bz", "BOL": "bo", "BOT": "bw", "BRB": "bb", "BRU": "bn", "BUL": "bg",
	"BHR": "bh", "CAM": "kh", "CAY": "ky", "CGO": "cg", "CHA": "td", "CHN": "cn",
	"COD": "cd", "COK": "ck", "COM": "km", "CPV": "cv", "CRC": "cr", "CTA": "cf",
	"CUB": "cu", "CUW": "cw", "CYP": "cy", "DJI": "dj", "DMA": "dm", "DOM": "do",
	"ECU": "ec", "EQG": "gq", "ERI": "er", "EST": "ee", "ETH": "et", "FIJ": "fj",
	"FIN": "fi", "FRO": "fo", "GAB": "ga", "GAM": "gm", "GEO": "ge", "GIB": "gi",
	"GNB": "gw", "GRN": "gd", "GUA": "gt", "GUI": "gn", "GUM": "gu", "GUY": "gy",
	"HAI": "ht", "HKG": "hk", "HON": "hn", "IDN": "id", "IND": "in", "IRN": "ir",
	"IRQ": "iq", "ISL": "is", "ISR": "il", "JAM": "jm", "JOR": "jo", "KAZ": "kz",
	"KEN": "ke", "KGZ": "kg", "KOS": "xk", "KSA": "sa", "KUW": "kw", "LAO": "la",
	"LBN": "lb", "LBR": "lr", "LBY": "ly", "LCA": "lc", "LES": "ls", "LIE": "li",
	"LTU": "lt", "LUX": "lu", "LVA": "lv", "MAC": "mo", "MAD": "mg", "MAS": "my",
	"MDA": "md", "MDV": "mv", "MGL": "mn", "MKD": "mk", "MLI": "ml", "MLT": "mt",
	"MNE": "me", "MOZ": "mz", "MRI": "mu", "MSR": "ms", "MTN": "mr", "MWI": "mw",
	"MYA": "mm", "NAM": "na", "NCA": "ni", "NCL": "nc", "NEP": "np", "NIG": "ne",
	"NZL": "nz", "OMA": "om", "PAK": "pk", "PAN": "pa", "PAR": "py", "PER": "pe",
	"PHI": "ph", "PLE": "ps", "PNG": "pg", "PRK": "kp", "PUR": "pr", "QAT": "qa",
	"RSA": "za", "RWA": "rw", "SAM": "ws", "SEY": "sc", "SGP": "sg",
	"SKN": "kn", "SLE": "sl", "SMR": "sm", "SOL": "sb", "SOM": "so", "SRI": "lk",
	"SSD": "ss", "STP": "st", "SUD": "sd", "SUR": "sr", "SVN": "si", "SWZ": "sz",
	"SYR": "sy", "TAH": "pf", "TAN": "tz", "TCA": "tc", "TGA": "to", "THA": "th",
	"TJK": "tj", "TKM": "tm", "TLS": "tl", "TOG": "tg", "TPE": "tw", "TRI": "tt",
	"UAE": "ae", "UGA": "ug", "UZB": "uz", "VAN": "vu", "VEN": "ve", "VGB": "vg",
	"VIE": "vn", "VIN": "vc", "VIR": "vi", "YEM": "ye", "ZAM": "zm", "ZIM": "zw",
	// Add more as needed, ensuring consistency with the FifaCountryCodes map
}
