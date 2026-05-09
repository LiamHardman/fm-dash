package main

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

const overridesConfigName = "club_logo_overrides.json"

// overridesMap maps normalized club name -> teamID.
// An empty string value means the user explicitly rejected any match for that name.
var (
	overridesMap    map[string]string
	overridesMutex  sync.RWMutex
	overridesLoaded bool
)

// LoadOverridesFromStorage loads overrides from the shared config storage on startup.
// Must be called after InitConfigStorage.
func LoadOverridesFromStorage() {
	overridesMutex.Lock()
	defer overridesMutex.Unlock()

	overridesMap = make(map[string]string)

	if configStorage == nil {
		LogWarn("Club logo overrides: config storage not initialized, starting with empty overrides")
		overridesLoaded = true
		return
	}

	data, err := configStorage.RetrieveConfig(overridesConfigName)
	if err != nil {
		if os.IsNotExist(err) {
			LogInfo("Club logo overrides: no overrides file found, starting with empty overrides")
		} else {
			LogWarn("Club logo overrides: failed to load overrides: %v", err)
		}
		overridesLoaded = true
		return
	}

	if err := json.Unmarshal(data, &overridesMap); err != nil {
		LogWarn("Club logo overrides: failed to parse overrides JSON: %v", err)
	}

	LogInfo("Club logo overrides: loaded %d overrides", len(overridesMap))
	overridesLoaded = true
}

func ensureOverridesLoaded() {
	overridesMutex.RLock()
	loaded := overridesLoaded
	overridesMutex.RUnlock()
	if !loaded {
		LoadOverridesFromStorage()
	}
}

// getLogoOverride looks up an override for the given raw club name.
// Returns (teamID, isRejection, found):
//   - found=false         → no override; proceed with normal matching
//   - found=true, isRejection=true  → user rejected any logo for this name
//   - found=true, isRejection=false → use teamID directly
func getLogoOverride(name string) (teamID string, isRejection bool, found bool) {
	ensureOverridesLoaded()
	key := normalizeOverrideName(name)

	overridesMutex.RLock()
	defer overridesMutex.RUnlock()

	val, ok := overridesMap[key]
	if !ok {
		return "", false, false
	}
	if val == "" {
		return "", true, true
	}
	return val, false, true
}

// SetLogoOverride saves an override. An empty teamID marks the name as rejected.
func SetLogoOverride(name, teamID string) error {
	ensureOverridesLoaded()
	key := normalizeOverrideName(name)

	overridesMutex.Lock()
	overridesMap[key] = teamID
	snapshot := snapshotOverrides()
	overridesMutex.Unlock()

	return persistOverrides(snapshot)
}

// DeleteLogoOverride removes any override for the given name.
func DeleteLogoOverride(name string) error {
	ensureOverridesLoaded()
	key := normalizeOverrideName(name)

	overridesMutex.Lock()
	delete(overridesMap, key)
	snapshot := snapshotOverrides()
	overridesMutex.Unlock()

	return persistOverrides(snapshot)
}

// ListLogoOverrides returns a copy of all overrides keyed by normalized name.
func ListLogoOverrides() map[string]string {
	ensureOverridesLoaded()

	overridesMutex.RLock()
	defer overridesMutex.RUnlock()

	return snapshotOverrides()
}

// snapshotOverrides returns a copy of overridesMap. Must be called with at least a read lock held.
func snapshotOverrides() map[string]string {
	out := make(map[string]string, len(overridesMap))
	for k, v := range overridesMap {
		out[k] = v
	}
	return out
}

func persistOverrides(data map[string]string) error {
	if configStorage == nil {
		LogWarn("Club logo overrides: config storage not initialized, override not persisted")
		return nil
	}
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return configStorage.StoreConfig(overridesConfigName, encoded)
}

func normalizeOverrideName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
