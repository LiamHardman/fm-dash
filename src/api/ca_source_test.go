package main

import "testing"

func TestSaveCAReplacesTheEstimate(t *testing.T) {
	estimated := saveOnlyRow(t, nil, nil)
	if estimated.CA == 0 || estimated.CA == 152 {
		t.Fatalf("estimate should be a non-trivial value other than 152, got %d", estimated.CA)
	}
	withSave := saveOnlyRow(t, []string{"CA", "PA"}, []string{"152", "171"})
	if withSave.CA != 152 {
		t.Errorf("CA = %d, want the save's 152", withSave.CA)
	}

	// Recalculation (weights change, cache rebuild) re-runs enhancement; CA must stay SaveCA.
	EnhancePlayerWithCalculations(&withSave)
	if withSave.CA != 152 {
		t.Errorf("after recalculation CA = %d, want 152", withSave.CA)
	}

	// MBR never reads CA.
	if withSave.MBR != estimated.MBR || withSave.Overall != estimated.Overall {
		t.Errorf("MBR/Overall changed with SaveCA: %d/%d vs %d/%d", withSave.MBR, withSave.Overall, estimated.MBR, estimated.Overall)
	}
}

func TestScoutReportStarsUseSaveCA(t *testing.T) {
	best := saveOnlyRow(t, []string{"CA"}, []string{"180"})
	subject := saveOnlyRow(t, []string{"CA"}, []string{"90"})
	if stars := computeStars(subject.CA, []Player{best}); stars != 2.5 {
		t.Errorf("stars = %v, want 2.5 (90/180 of 5)", stars)
	}
}
