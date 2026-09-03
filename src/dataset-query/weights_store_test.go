package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestSeedDefaultsIfEmpty(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	if err := store.seedDefaultsIfEmpty(ctx); err != nil {
		t.Fatalf("seedDefaultsIfEmpty failed: %v", err)
	}

	got, err := store.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if !reflect.DeepEqual(got, defaultAttributeWeights) {
		t.Errorf("seeded weights don't match defaultAttributeWeights\ngot:  %v\nwant: %v", got, defaultAttributeWeights)
	}
}

func TestSeedSkipsIfAlreadyPopulated(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	if err := store.seedDefaultsIfEmpty(ctx); err != nil {
		t.Fatalf("first seedDefaultsIfEmpty failed: %v", err)
	}

	// Manually override PAC, then seed again - the override must survive.
	if err := store.SetCategories(ctx, map[string]map[string]int{"PAC": {"Acc": 42}}); err != nil {
		t.Fatalf("SetCategories override failed: %v", err)
	}
	if err := store.seedDefaultsIfEmpty(ctx); err != nil {
		t.Fatalf("second seedDefaultsIfEmpty failed: %v", err)
	}

	got, err := store.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	want := map[string]int{"Acc": 42}
	if !reflect.DeepEqual(got["PAC"], want) {
		t.Errorf("PAC override was clobbered by re-seed: got %v, want %v", got["PAC"], want)
	}
}

func TestSetCategoriesReplaceSemantics(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store, err := openWeightsStore(ctx, dir)
	if err != nil {
		t.Fatalf("openWeightsStore failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	if err := store.SetCategories(ctx, map[string]map[string]int{"PAC": {"Acc": 3}}); err != nil {
		t.Fatalf("SetCategories(PAC) failed: %v", err)
	}
	if err := store.SetCategories(ctx, map[string]map[string]int{"SHO": {"Fin": 7}}); err != nil {
		t.Fatalf("SetCategories(SHO) failed: %v", err)
	}

	got, err := store.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if !reflect.DeepEqual(got["PAC"], map[string]int{"Acc": 3}) {
		t.Errorf("PAC was disturbed by an unrelated SHO update: got %v", got["PAC"])
	}
	if !reflect.DeepEqual(got["SHO"], map[string]int{"Fin": 7}) {
		t.Errorf("SHO not set correctly: got %v", got["SHO"])
	}

	// Wholesale replace: a new PAC update must fully replace the old one,
	// not merge into it.
	if err := store.SetCategories(ctx, map[string]map[string]int{"PAC": {"Pac": 9}}); err != nil {
		t.Fatalf("SetCategories(PAC) replace failed: %v", err)
	}
	got, err = store.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if !reflect.DeepEqual(got["PAC"], map[string]int{"Pac": 9}) {
		t.Errorf("PAC was merged instead of wholesale-replaced: got %v, want map[Pac:9]", got["PAC"])
	}
}

func TestValidateWeightsUpdateRejectsOutOfRange(t *testing.T) {
	if err := validateWeightsUpdate(map[string]map[string]int{"PAC": {"Acc": -1}}); err == nil {
		t.Error("expected error for weight -1, got nil")
	} else if !errors.As(err, new(*WeightsValidationError)) {
		t.Errorf("expected *WeightsValidationError, got %T", err)
	}

	if err := validateWeightsUpdate(map[string]map[string]int{"PAC": {"Acc": maxAttributeWeightValue + 1}}); err == nil {
		t.Error("expected error for weight above max, got nil")
	}

	// Inclusive bounds must be accepted.
	if err := validateWeightsUpdate(map[string]map[string]int{"PAC": {"Acc": 0}}); err != nil {
		t.Errorf("weight 0 should be valid, got error: %v", err)
	}
	if err := validateWeightsUpdate(map[string]map[string]int{"PAC": {"Acc": maxAttributeWeightValue}}); err != nil {
		t.Errorf("weight %d should be valid, got error: %v", maxAttributeWeightValue, err)
	}
}

func TestValidateWeightsUpdateRejectsEmpty(t *testing.T) {
	if err := validateWeightsUpdate(map[string]map[string]int{}); err == nil {
		t.Error("expected error for empty top-level map, got nil")
	}
	if err := validateWeightsUpdate(map[string]map[string]int{"PAC": {}}); err == nil {
		t.Error("expected error for empty category, got nil")
	}
}
