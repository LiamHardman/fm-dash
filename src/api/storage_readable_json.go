package main

// The encoding preference governs new writes, not which existing Datasets can
// be read. This keeps a protobuf-to-JSON configuration rollback lossless.
type readableJSONStorage struct{ StorageInterface }

func (s readableJSONStorage) Retrieve(id string) (DatasetData, error) {
	return CreateProtobufStorage(s.StorageInterface).Retrieve(id)
}
