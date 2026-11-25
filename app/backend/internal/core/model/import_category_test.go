package model

import "testing"

func TestAfterImportCategoryFor(t *testing.T) {
	d := Downloader{CategoryAfterImport: "shared", CategoryAfterImportMovies: "movies-imported", CategoryAfterImportShows: "custom TV", CategoryAfterImportMusic: "albums"}
	for _, tc := range []struct{ kind, want string }{{"movie", "movies-imported"}, {"show", "custom TV"}, {"artist", "albums"}, {"", "shared"}} {
		if got := d.AfterImportCategoryFor(tc.kind); got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.kind, got, tc.want)
		}
	}
	d.CategoryAfterImportMovies = ""
	if got := d.AfterImportCategoryFor("movie"); got != "shared" {
		t.Fatalf("fallback %q", got)
	}
	d.CategoryAfterImport = ""
	if got := d.AfterImportCategoryFor("movie"); got != "" {
		t.Fatalf("unset %q", got)
	}
	if got := d.AfterImportCategoryFor("show"); got != "custom TV" {
		t.Fatalf("override without shared %q", got)
	}
}
