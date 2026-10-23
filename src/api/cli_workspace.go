package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type cliDataset struct {
	ID                  string         `json:"datasetId"`
	Label               string         `json:"label"`
	Created             time.Time      `json:"created"`
	Expires             time.Time      `json:"expires"`
	InputSHA256         string         `json:"inputSha256"`
	ConfigurationSHA256 string         `json:"configurationSha256"`
	Result              *ParsedPlayers `json:"result"`
}

type cliWorkspace struct{ dir string }

func cliOpenWorkspace(path string) (*cliWorkspace, error) {
	if path == "" {
		root, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(root, "fmdash")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return &cliWorkspace{abs}, nil
}

func (w *cliWorkspace) datasetPath(id string) (string, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", cliFail(2, "--dataset must be a UUID")
	}
	return filepath.Join(w.dir, "datasets", id+".json"), nil
}

func (w *cliWorkspace) readDataset(ctx context.Context, id string) (*cliDataset, error) {
	path, err := w.datasetPath(id)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, cliFail(5, "Dataset %s is unavailable", id)
	}
	if err != nil {
		return nil, cliFail(6, "Dataset read: %v", err)
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	var d cliDataset
	if err = cliDecodeJSON(data, &d); err != nil {
		return nil, err
	}
	if d.Result == nil {
		return nil, cliFail(3, "invalid Dataset artifact")
	}
	if !d.Expires.IsZero() && time.Now().After(d.Expires) {
		return nil, cliFail(5, "Dataset %s has expired", id)
	}
	for i := range d.Result.Players {
		if i < len(d.Result.RawAttributes) {
			d.Result.Players[i].Attributes = d.Result.RawAttributes[i]
		}
	}
	return &d, nil
}

func (w *cliWorkspace) listDatasets(ctx context.Context) ([]map[string]any, error) {
	entries, err := os.ReadDir(filepath.Join(w.dir, "datasets"))
	if os.IsNotExist(err) {
		return []map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	items := []map[string]any{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		path, e := w.datasetPath(id)
		if e != nil {
			continue
		}
		data, e := os.ReadFile(path)
		if e != nil {
			return nil, e
		}
		var d cliDataset
		if e = json.Unmarshal(data, &d); e != nil {
			return nil, e
		}
		available := d.Expires.IsZero() || time.Now().Before(d.Expires)
		count := 0
		if d.Result != nil {
			count = len(d.Result.Players)
		}
		items = append(items, map[string]any{"datasetId": d.ID, "label": d.Label, "created": d.Created, "expires": d.Expires, "available": available, "playerCount": count, "inputSha256": d.InputSHA256})
		if e = ctx.Err(); e != nil {
			return nil, e
		}
	}
	return items, nil
}

// All workspace mutations share one process-independent OS lock. The callback
// rereads state after acquiring the lock, so imports/renames/restores cannot race.
func (w *cliWorkspace) mutate(ctx context.Context, fn func() error) error {
	if err := os.MkdirAll(w.dir, 0o750); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(w.dir, "workspace.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	for {
		locked, lockErr := tryLockConfigFile(f)
		if lockErr != nil {
			return lockErr
		}
		if locked {
			break
		}
		timer := time.NewTimer(20 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	defer unlockConfigFile(f)
	if err = ctx.Err(); err != nil {
		return err
	}
	return fn()
}

func (w *cliWorkspace) writeDataset(d *cliDataset) error {
	path, err := w.datasetPath(d.ID)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return atomicConfigWrite(path, data)
}

func (w *cliWorkspace) readProfile() (ProfileBackup, string, error) {
	path := filepath.Join(w.dir, "profile", "backup.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ProfileBackup{1, defaultShortlistDocument(), SavedSearchDocument{1, []SavedSearch{}}, ProgressionProjectDocument{1, []ProgressionProject{}}}, profileRevision(nil), nil
	}
	if err != nil {
		return ProfileBackup{}, "", err
	}
	var backup ProfileBackup
	if err = cliDecodeJSON(data, &backup); err != nil {
		return backup, "", err
	}
	return backup, profileRevision(data), nil
}

func validateCLIBackup(backup *ProfileBackup) error {
	if backup.Version != 1 {
		return cliFail(3, "backup version must be 1")
	}
	if err := validateShortlistDocument(&backup.Shortlists); err != nil {
		return cliFail(3, "shortlists: %v", err)
	}
	if err := validateSavedSearchDocument(&backup.SavedSearches); err != nil {
		return cliFail(3, "saved searches: %v", err)
	}
	if !validateProgressionProjects(&backup.ProgressionProjects) {
		return cliFail(3, "invalid Progression project document")
	}
	for _, doc := range []any{backup.Shortlists, backup.SavedSearches, backup.ProgressionProjects} {
		data, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		if len(data) > 2<<20 {
			return cliFail(3, "profile document exceeds 2 MiB")
		}
	}
	return nil
}

func (w *cliWorkspace) writeProfile(backup *ProfileBackup) (string, error) {
	if err := validateCLIBackup(backup); err != nil {
		return "", err
	}
	data, err := json.Marshal(backup)
	if err != nil {
		return "", err
	}
	path := filepath.Join(w.dir, "profile", "backup.json")
	if err = os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return "", err
	}
	return profileRevision(data), atomicConfigWrite(path, data)
}

func cliProfile(ctx context.Context, command string, o cliOptions, stdin io.Reader, w *cliWorkspace) (any, error) {
	if command == "recipes encode" || command == "recipes decode" {
		if command == "recipes decode" {
			data, err := base64.RawURLEncoding.DecodeString(o.query)
			if err != nil {
				return nil, cliFail(3, "invalid Search recipe")
			}
			var recipe struct {
				Version int             `json:"version"`
				Filters json.RawMessage `json:"filters"`
			}
			if err = cliDecodeJSON(data, &recipe); err != nil {
				return nil, err
			}
			if recipe.Version != 1 || containsSavedSearchForbiddenData(recipe.Filters) {
				return nil, cliFail(3, "unsupported Search recipe")
			}
			var filters map[string]any
			if err = cliDecodeJSON(recipe.Filters, &filters); err != nil {
				return nil, err
			}
			return filters, nil
		}
		data, err := cliBody(ctx, o, stdin)
		if err != nil {
			return nil, err
		}
		var filters map[string]any
		if err = cliDecodeJSON(data, &filters); err != nil {
			return nil, err
		}
		if filters == nil || containsSavedSearchForbiddenData(data) {
			return nil, cliFail(3, "Search recipes contain criteria only")
		}
		encoded, _ := json.Marshal(map[string]any{"version": 1, "filters": filters})
		return map[string]any{"recipe": base64.RawURLEncoding.EncodeToString(encoded)}, nil
	}
	if command == "profile validate" {
		data, err := cliBody(ctx, o, stdin)
		if err != nil {
			return nil, err
		}
		var b ProfileBackup
		if err = cliDecodeJSON(data, &b); err != nil {
			return nil, err
		}
		return b, validateCLIBackup(&b)
	}
	var output any
	err := w.mutate(ctx, func() error {
		backup, revision, err := w.readProfile()
		if err != nil {
			return err
		}
		switch command {
		case "profile export":
			output = backup
			return nil
		case "shortlists list":
			output = map[string]any{"document": backup.Shortlists, "revision": revision}
			return nil
		case "saved-searches list":
			output = map[string]any{"document": backup.SavedSearches, "revision": revision}
			return nil
		case "progression projects":
			output = map[string]any{"document": backup.ProgressionProjects, "revision": revision}
			return nil
		}
		data, err := cliBody(ctx, o, stdin)
		if err != nil {
			return err
		}
		if command == "profile restore" {
			var incoming ProfileBackup
			if err = cliDecodeJSON(data, &incoming); err != nil {
				return err
			}
			if err = validateCLIBackup(&incoming); err != nil {
				return err
			}
			// Mirrors mergeProfileBackup: retain current order, append additions;
			// replacing an ID retains its existing location in the document.
			changes := []map[string]string{}
			merge := func(section string, current, incoming any) any {
				ca, _ := json.Marshal(current)
				ia, _ := json.Marshal(incoming)
				var a, b []map[string]any
				_ = json.Unmarshal(ca, &a)
				_ = json.Unmarshal(ia, &b)
				index := map[string]int{}
				for i, v := range a {
					index[v["id"].(string)] = i
				}
				for _, v := range b {
					id := v["id"].(string)
					i, exists := index[id]
					action := "Add"
					if exists {
						action = "Keep existing"
						if o.replace {
							action = "Replace existing"
							a[i] = v
						}
					} else {
						index[id] = len(a)
						a = append(a, v)
					}
					changes = append(changes, map[string]string{"section": section, "id": id, "name": v["name"].(string), "action": action})
				}
				return a
			}
			decode := func(value, target any) error {
				b, e := json.Marshal(value)
				if e != nil {
					return e
				}
				return json.Unmarshal(b, target)
			}
			if err = decode(merge("shortlists", backup.Shortlists.Lists, incoming.Shortlists.Lists), &backup.Shortlists.Lists); err != nil {
				return err
			}
			if err = decode(merge("savedSearches", backup.SavedSearches.Searches, incoming.SavedSearches.Searches), &backup.SavedSearches.Searches); err != nil {
				return err
			}
			if err = decode(merge("progressionProjects", backup.ProgressionProjects.Projects, incoming.ProgressionProjects.Projects), &backup.ProgressionProjects.Projects); err != nil {
				return err
			}
			if err = validateCLIBackup(&backup); err != nil {
				return err
			}
			output = map[string]any{"document": backup, "changes": changes, "preview": o.dryRun, "previousRevision": revision}
			if o.dryRun {
				return nil
			}
		} else {
			if o.id == "" {
				return cliFail(2, "document replacement requires --id with the revision returned by list")
			}
			if o.id != revision {
				return cliFail(7, "profile revision conflict; reread and retry")
			}
			switch command {
			case "shortlists set":
				err = cliDecodeJSON(data, &backup.Shortlists)
			case "saved-searches set":
				err = cliDecodeJSON(data, &backup.SavedSearches)
			case "progression set":
				err = cliDecodeJSON(data, &backup.ProgressionProjects)
			default:
				return cliFail(2, "unknown profile operation")
			}
			if err != nil {
				return err
			}
		}
		newRevision, err := w.writeProfile(&backup)
		if err != nil {
			return err
		}
		if output == nil {
			output = map[string]any{"revision": newRevision, "document": backup}
		}
		return nil
	})
	return output, err
}

func cliImport(ctx context.Context, args []string, o cliOptions, stdin io.Reader, w *cliWorkspace, e *cliEnvelope) (any, error) {
	if len(args) != 1 {
		return nil, cliFail(2, "datasets import requires one CSV/HTML input")
	}
	format, err := cliInputFormat(args[0], o.inputFormat)
	if err != nil {
		return nil, err
	}
	data, err := cliReadInput(ctx, args[0], stdin, o.maxBytes)
	if err != nil {
		return nil, err
	}
	parsed, err := ParsePlayers(ctx, bytes.NewReader(data), ParseOptions{format, o.stage, o.workers, o.strict})
	if err != nil {
		return parsed, err
	}
	e.Diagnostics = parsed.Diagnostics
	if parsed.Counts.Errors > 0 {
		e.Status = "partial"
	}
	d := &cliDataset{ID: uuid.NewString(), Label: o.name, Created: time.Now().UTC(), InputSHA256: cliHash(data), ConfigurationSHA256: cliString(e.Provenance["configurationSha256"]), Result: parsed}
	d.Expires = d.Created.Add(getRetentionPeriod())
	if d.Label == "" {
		d.Label = filepath.Base(args[0])
	}
	err = w.mutate(ctx, func() error { return w.writeDataset(d) })
	return d, err
}

func cliReferences(ctx context.Context, o cliOptions, stdin io.Reader, w *cliWorkspace) (any, error) {
	data, err := cliBody(ctx, o, stdin)
	if err != nil {
		return nil, err
	}
	var req struct {
		References []ShortlistPlayerRef `json:"references"`
	}
	if err = cliDecodeJSON(data, &req); err != nil {
		return nil, err
	}
	if len(req.References) > 2000 {
		return nil, cliFail(3, "at most 2000 Player references are allowed")
	}
	items := []resolvedPlayerReference{}
	for _, ref := range req.References {
		if !validPlayerRef(ref) {
			return nil, cliFail(3, "invalid Player reference")
		}
		item := resolvedPlayerReference{PlayerRef: ref, State: "missing"}
		d, err := w.readDataset(ctx, ref.DatasetID)
		if err != nil {
			var failure *cliError
			if errors.As(err, &failure) && failure.code == 5 {
				item.State = "expired"
			} else {
				return nil, err
			}
		} else {
			for _, p := range d.Result.Players {
				if p.UID == ref.PlayerUID {
					item.State = "available"
					item.Player = &referencedPlayer{p.UID, p.Name, p.Club, p.Position, p.ParsedPositions}
					break
				}
			}
		}
		items = append(items, item)
	}
	return map[string]any{"items": items}, nil
}

func cliSortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
