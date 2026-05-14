package main

import (
	"context"
	"hash/fnv"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/minio/minio-go/v7"
)

// fifaToRegenRegion maps FIFA 3-letter country codes to regen folder names.
var fifaToRegenRegion = map[string]string{
	// African (Sub-Saharan Africa)
	"NGA": "African", "GHA": "African", "SEN": "African", "CIV": "African",
	"CMR": "African", "MLI": "African", "GAM": "African", "GNB": "African",
	"BEN": "African", "TOG": "African", "NIG": "African", "SLE": "African",
	"LBR": "African", "GUI": "African", "EQG": "African", "GAB": "African",
	"CGO": "African", "COD": "African", "ANG": "African", "MOZ": "African",
	"ZIM": "African", "ZAM": "African", "MWI": "African", "TAN": "African",
	"KEN": "African", "UGA": "African", "ETH": "African", "ERI": "African",
	"SSD": "African", "RWA": "African", "BDI": "African", "MAD": "African",
	"COM": "African", "SEY": "African", "MRI": "African", "STP": "African",
	"CPV": "African", "NAM": "African", "BOT": "African", "RSA": "African",
	"SWZ": "African", "LES": "African", "BFA": "African", "CHA": "African",
	"CTA": "African", "DJI": "African", "SOM": "African", "SUD": "African",

	// Asian (East Asia)
	"JPN": "Asian", "KOR": "Asian", "CHN": "Asian", "PRK": "Asian",
	"TPE": "Asian", "HKG": "Asian", "MGL": "Asian", "MAC": "Asian",

	// Caucasian (Western/Anglophone European)
	"ENG": "Caucasian", "GER": "Caucasian", "FRA": "Caucasian", "NED": "Caucasian",
	"BEL": "Caucasian", "AUT": "Caucasian", "SUI": "Caucasian", "IRL": "Caucasian",
	"SCO": "Caucasian", "WAL": "Caucasian", "NIR": "Caucasian", "AUS": "Caucasian",
	"NZL": "Caucasian", "CAN": "Caucasian", "USA": "Caucasian", "LUX": "Caucasian",
	"LIE": "Caucasian", "ISR": "Caucasian",

	// Central European
	"POL": "Central European", "CZE": "Central European", "SVK": "Central European",
	"HUN": "Central European", "LTU": "Central European", "LVA": "Central European",
	"EST": "Central European",

	// EECA (Eastern Europe & Central Asia)
	"RUS": "EECA", "UKR": "EECA", "BLR": "EECA", "KAZ": "EECA",
	"UZB": "EECA", "KGZ": "EECA", "TJK": "EECA", "TKM": "EECA",
	"GEO": "EECA", "ARM": "EECA", "AZE": "EECA", "MDA": "EECA",

	// Italmed (Italian / close Mediterranean)
	"ITA": "Italmed", "MLT": "Italmed", "CYP": "Italmed", "SMR": "Italmed",

	// MENA (Middle East & North Africa)
	"EGY": "MENA", "MAR": "MENA", "ALG": "MENA", "TUN": "MENA",
	"LBY": "MENA", "JOR": "MENA", "LBN": "MENA", "SYR": "MENA",
	"IRQ": "MENA", "IRN": "MENA", "KSA": "MENA", "UAE": "MENA",
	"QAT": "MENA", "KUW": "MENA", "OMA": "MENA", "BHR": "MENA",
	"YEM": "MENA", "PLE": "MENA", "MTN": "MENA",

	// MESA (South Asia)
	"IND": "MESA", "PAK": "MESA", "BAN": "MESA", "SRI": "MESA",
	"NEP": "MESA", "BHU": "MESA", "MDV": "MESA", "AFG": "MESA",

	// SAMed (Southern-cone South America — European-descended)
	"ARG": "SAMed", "URU": "SAMed",

	// Scandivavian (preserves the folder's original spelling)
	"SWE": "Scandivavian", "NOR": "Scandivavian", "DEN": "Scandivavian",
	"FIN": "Scandivavian", "ISL": "Scandivavian", "FRO": "Scandivavian",

	// Seasian (Southeast Asia)
	"THA": "Seasian", "VIE": "Seasian", "MYA": "Seasian", "IDN": "Seasian",
	"MAS": "Seasian", "PHI": "Seasian", "SGP": "Seasian", "CAM": "Seasian",
	"LAO": "Seasian", "TLS": "Seasian", "BRU": "Seasian",

	// South American (Latin/South America)
	"BRA": "South American", "COL": "South American", "CHI": "South American",
	"ECU": "South American", "PAR": "South American", "BOL": "South American",
	"PER": "South American", "VEN": "South American", "GUY": "South American",
	"SUR": "South American", "TRI": "South American", "JAM": "South American",
	"HAI": "South American", "HON": "South American", "GUA": "South American",
	"NCA": "South American", "CRC": "South American", "PAN": "South American",
	"BLZ": "South American",

	// SpanMed (Spanish/Portuguese/Mediterranean)
	"ESP": "SpanMed", "POR": "SpanMed", "MEX": "SpanMed",
	"DOM": "SpanMed", "PUR": "SpanMed", "CUB": "SpanMed",

	// YugoGreek (Balkans / Yugoslav / Greek region)
	"SRB": "YugoGreek", "CRO": "YugoGreek", "BIH": "YugoGreek",
	"MNE": "YugoGreek", "SVN": "YugoGreek", "KOS": "YugoGreek",
	"GRE": "YugoGreek", "ALB": "YugoGreek", "BUL": "YugoGreek",
	"ROU": "YugoGreek", "MKD": "YugoGreek",
}

// regenImageCache holds sorted lists of .png filenames per region, populated at startup.
var regenImageCache = map[string][]string{}

// regenLocalBaseDir is set when the cache was loaded from local filesystem.
var regenLocalBaseDir string

// initRegenCache populates regenImageCache from S3 (if configured) or the local filesystem.
func initRegenCache() {
	if s3 := underlyingS3Storage(); s3 != nil && s3.client != nil {
		initRegenCacheFromS3(s3)
		return
	}
	initRegenCacheFromLocal(getRegenDirectory())
}

// initRegenCacheFromS3 lists objects under the regen/ prefix in the same bucket as faces.
func initRegenCacheFromS3(s3 *S3Storage) {
	bucket := regenS3Bucket(s3)
	prefix := "regen/"

	ctx := context.Background()
	objectCh := s3.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	count := 0
	for obj := range objectCh {
		if obj.Err != nil {
			LogWarn("Error listing regen objects from S3: %v", obj.Err)
			continue
		}
		// Key format: "regen/African/KT5_African1.png"
		rel := strings.TrimPrefix(obj.Key, prefix)
		parts := strings.SplitN(rel, "/", 2)
		if len(parts) != 2 || parts[1] == "" {
			continue
		}
		region, filename := parts[0], parts[1]
		if !strings.EqualFold(filepath.Ext(filename), ".png") {
			continue
		}
		regenImageCache[region] = append(regenImageCache[region], filename)
		count++
	}

	for region := range regenImageCache {
		sort.Strings(regenImageCache[region])
	}

	LogInfo("Regen cache initialised from S3: %d regions, %d images", len(regenImageCache), count)
}

// initRegenCacheFromLocal scans the local regen directory.
func initRegenCacheFromLocal(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		LogWarn("Regen directory not found or unreadable, fallback faces disabled: %v", err)
		return
	}

	regenLocalBaseDir = dir
	count := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		region := entry.Name()
		files, err := os.ReadDir(filepath.Join(dir, region))
		if err != nil {
			LogWarn("Could not read regen region folder %s: %v", region, err)
			continue
		}
		for _, f := range files {
			if !f.IsDir() && strings.EqualFold(filepath.Ext(f.Name()), ".png") {
				regenImageCache[region] = append(regenImageCache[region], f.Name())
				count++
			}
		}
		sort.Strings(regenImageCache[region])
	}

	LogInfo("Regen cache initialised from local disk: %d regions, %d images", len(regenImageCache), count)
}

// getRegenDirectory returns the local regen directory path.
func getRegenDirectory() string {
	if d := os.Getenv("REGEN_DIR"); d != "" {
		return d
	}
	return "./regen"
}

// regenS3Bucket returns the S3 bucket to use for regen images (same as faces).
func regenS3Bucket(s3 *S3Storage) string {
	if b := os.Getenv("S3_FACES_BUCKET"); b != "" {
		return b
	}
	return s3.bucketName
}

// getRegenFileInfo resolves a player UID + FIFA code to a (region, filename) pair.
// Uses uid % imageCount for deterministic, stateless selection.
func getRegenFileInfo(uidStr, fifaCode string) (region, filename string, ok bool) {
	region, ok = fifaToRegenRegion[strings.ToUpper(fifaCode)]
	if !ok {
		return "", "", false
	}

	images := regenImageCache[region]
	if len(images) == 0 {
		return "", "", false
	}

	// Parse UID to int64; fall back to FNV hash for non-numeric UIDs.
	var index int64
	parsed := int64(0)
	valid := true
	for _, ch := range uidStr {
		if ch < '0' || ch > '9' {
			valid = false
			break
		}
		parsed = parsed*10 + int64(ch-'0')
	}
	if valid && uidStr != "" {
		index = parsed
	} else {
		h := fnv.New64a()
		h.Write([]byte(uidStr))
		index = int64(h.Sum64())
	}

	n := int64(len(images))
	idx := index % n
	if idx < 0 {
		idx += n
	}

	return region, images[idx], true
}

// serveRegenFace serves a regen image for the given player UID and FIFA code.
// Tries S3 first (if configured), then the local filesystem.
func serveRegenFace(ctx context.Context, w http.ResponseWriter, r *http.Request, uid, fifaCode string) bool {
	region, filename, ok := getRegenFileInfo(uid, fifaCode)
	if !ok {
		return false
	}

	if s3 := underlyingS3Storage(); s3 != nil && s3.client != nil {
		err := s3.getRegenImage(ctx, region, filename, w)
		if err == nil {
			return true
		}
		LogWarn("Failed to serve regen image from S3, trying local: region=%s file=%s err=%v", region, filename, err)
	}

	if regenLocalBaseDir != "" {
		localPath := filepath.Join(regenLocalBaseDir, region, filename)
		if _, err := os.Stat(localPath); err == nil {
			http.ServeFile(w, r, localPath)
			return true
		}
	}

	return false
}
