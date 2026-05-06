package main

import (
	"context"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	configInitOnce.Do(initializeConfigAsync)
	if err := EnsureConfigInitialized(configLoadTimeout); err != nil {
		LogWarn("Test config initialization completed with error: %v", err)
	}

	muAttributeWeights.RLock()
	globalAttributeWeights = deepCopyWeights(attributeWeights)
	muAttributeWeights.RUnlock()

	if storage == nil {
		storage = InitializeStorage(context.Background())
	}

	os.Exit(m.Run())
}
