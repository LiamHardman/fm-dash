package main

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestScanExportStyledHTMLPreservesCellText(t *testing.T) {
	input := `<table></table><table class="export"><tr><th title="Name">Name</th><th>Club</th></tr>` +
		`<tr><td style="color:red" data-note="ignored &amp; styling">  Jos&eacute; <span>Garc&iacute;a</span> </td>` +
		`<td class="club">A &amp; B<br>United &#163;</td></tr></table>` +
		`<table><tr><td>Must not enter the Dataset</td></tr></table>`
	var rows [][]string
	err := scanExport(context.Background(), strings.NewReader(input), "html", func(row int, cells []string, header bool) error {
		if row != len(rows)+1 || header != (row == 1) {
			t.Fatalf("row identity: row=%d header=%v", row, header)
		}
		rows = append(rows, cells)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"Name", "Club"}, {"José García", "A & B United £"}}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("cell text changed: got %q, want %q", rows, want)
	}
}
