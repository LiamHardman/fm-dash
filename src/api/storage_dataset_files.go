package main

import "strings"

const (
	datasetObjectPrefix = "datasets/"
	datasetExtProtobuf  = ".pb.gz"
	datasetExtJSONGzip  = ".json.gz"
	datasetExtJSON      = ".json"
)

var datasetFileExtensions = []string{
	datasetExtProtobuf,
	datasetExtJSONGzip,
	datasetExtJSON,
}

func datasetIDFromStorageName(name string) (string, bool) {
	id := strings.TrimPrefix(name, datasetObjectPrefix)
	for _, ext := range datasetFileExtensions {
		if strings.HasSuffix(id, ext) {
			return strings.TrimSuffix(id, ext), true
		}
	}
	return "", false
}

func datasetFormatFromStorageName(name string) string {
	switch {
	case strings.HasSuffix(name, datasetExtProtobuf):
		return "protobuf_gzip"
	case strings.HasSuffix(name, datasetExtJSONGzip):
		return "json_gzip"
	case strings.HasSuffix(name, datasetExtJSON):
		return "json"
	default:
		return "unknown"
	}
}

func appendDatasetIDOnce(ids []string, seen map[string]struct{}, id string) []string {
	if _, exists := seen[id]; exists {
		return ids
	}
	seen[id] = struct{}{}
	return append(ids, id)
}
