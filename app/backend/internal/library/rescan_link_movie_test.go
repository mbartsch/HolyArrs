package library

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HolyArrs/holyarrs/app/backend/internal/core/database"
	domain "github.com/HolyArrs/holyarrs/app/backend/internal/core/model"
	"github.com/HolyArrs/holyarrs/app/backend/internal/mediafiles"
)

// movieWithFoundCopy adds a movie whose slot has no file, ties it to a
// download in the given state (none when state is empty) and puts a video
// in its folder, as another app that imported the download does.
func movieWithFoundCopy(t *testing.T, state string) (*Service, *database.Store, int64, string) {
	t.Helper()
	s := open(t)
	ctx := context.Background()
	lf := addFolder(t, s)
	id, err := AddMovie(ctx, s, dune, AddOptions{}, nil, now, domain.System)
	if err != nil {
		t.Fatal(err)
	}
	v := versionsOf(t, s, id)[0]
	slotID := slotsOf(t, s, id)[0].ID
	if state != "" {
		err = s.Write(ctx, func(tx *sql.Tx) error {
			d := &domain.Downloader{Name: "q", Adapter: "qbittorrent", Protocol: "torrent", URL: "http://q", Category: "holyarrs", PathMap: []domain.PathMap{}, AfterImport: "leave", State: "ok"}
			if _, err := database.Insert(tx, "downloaders", d, domain.System); err != nil {
				return err
			}
			dl := &domain.Download{DownloaderID: d.ID, SlotIDs: []int64{slotID}, ExternalID: "abc", Progress: 1, State: state, FilesSeen: []domain.SeenFile{},
				SavePath: filepath.Join(lf.Path, "gone"), StartedBy: domain.System}
			if _, err := database.Insert(tx, "downloads", dl, domain.System); err != nil {
				return err
			}
			_, err := tx.Exec("UPDATE slots SET state = 'needs_a_decision', active_download_id = ? WHERE id = ?", dl.ID, slotID)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	rel := v.FolderName + "/Dune.Part.Two.2024.1080p.WEB-DL.H264-GRP.mkv"
	writeFile(t, filepath.Join(lf.Path, filepath.FromSlash(rel)), "placed by another app")
	running := func(tx *sql.Tx, dlID int64) (bool, error) {
		var st string
		if err := tx.QueryRow("SELECT state FROM downloads WHERE id = ?", dlID).Scan(&st); err != nil {
			return false, err
		}
		return st == "downloading", nil
	}
	return &Service{Store: s, Now: func() time.Time { return now }, DownloadRunning: running}, s, id, rel
}

// A container change can leave the old MP4 missing beside a new MKV and
// another cut. Only an unambiguous match to the old path may replace it.
func TestRescanMovieReplacement(t *testing.T) {
	for _, tc := range []struct {
		name            string
		oldName         string
		extra           []string
		protected       bool
		keepOld         bool
		wantReplacement bool
	}{
		{name: "one replacement", wantReplacement: true},
		{name: "matching stem beside another cut", extra: []string{"Other.Cut.mkv"}, wantReplacement: true},
		{name: "two matching containers", extra: []string{"Dune.Part.Two.2024.1080p.WEB-DL.H264-GRP.avi"}},
		{name: "no matching stem", oldName: "Old.Name.mp4", extra: []string{"Other.Cut.mkv"}},
		{name: "protected missing copy", extra: []string{"Other.Cut.mkv"}, protected: true},
		{name: "present copy", extra: []string{"Other.Cut.mkv"}, keepOld: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, s, id, replacement := movieWithFoundCopy(t, "")
			ctx := context.Background()
			v := versionsOf(t, s, id)[0]
			var lf mediafiles.LibraryFolder
			if err := database.Get(ctx, s.Read(), "library_folders", v.LibraryFolderID, &lf); err != nil {
				t.Fatal(err)
			}
			oldPath := replacement[:len(replacement)-len(filepath.Ext(replacement))] + ".mp4"
			if tc.oldName != "" {
				oldPath = v.FolderName + "/" + tc.oldName
			}
			writeFile(t, filepath.Join(lf.Path, oldPath), "original video")
			old, err := svc.AdoptMovieCopy(ctx, id, v.ID, CopyAdoption{RelativePath: oldPath}, domain.System)
			if err != nil {
				t.Fatal(err)
			}
			if tc.protected {
				if err := s.Write(ctx, func(tx *sql.Tx) error {
					_, err := tx.Exec("UPDATE files SET protected = 1 WHERE id = ?", old.ID)
					return err
				}); err != nil {
					t.Fatal(err)
				}
			}
			if !tc.keepOld {
				if err := os.Remove(filepath.Join(lf.Path, oldPath)); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range tc.extra {
				writeFile(t, filepath.Join(lf.Path, v.FolderName, name), "another video")
			}
			for pass := 0; pass < 2; pass++ {
				if _, err := svc.Rescan(ctx, domain.System); err != nil {
					t.Fatal(err)
				}
				sl := slotsOf(t, s, id)[0]
				if sl.FileID == nil {
					t.Fatal("movie lost its file link")
				}
				var linked mediafiles.File
				if err := database.Get(ctx, s.Read(), "files", *sl.FileID, &linked); err != nil {
					t.Fatal(err)
				}
				if tc.wantReplacement {
					if linked.RelativePath != replacement || linked.State != "present" || sl.State != mediafiles.SlotInLibrary {
						t.Fatalf("replacement not linked: slot=%+v file=%+v", sl, linked)
					}
					var previous mediafiles.File
					if err := database.Get(ctx, s.Read(), "files", old.ID, &previous); err != nil {
						t.Fatal(err)
					}
					if previous.State != "missing" || len(previous.SlotIDs) != 0 {
						t.Fatalf("old copy still owns slot: %+v", previous)
					}
				} else if linked.ID != old.ID {
					t.Fatalf("replaced a copy requiring a choice: %+v", linked)
				}
				if _, err := os.Stat(filepath.Join(lf.Path, replacement)); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func linkedCopy(t *testing.T, s *database.Store, titleID int64, rel string) (mediafiles.Slot, string) {
	t.Helper()
	sl := slotsOf(t, s, titleID)[0]
	var origin string
	_ = s.Read().QueryRow("SELECT origin FROM files WHERE relative_path = ?", rel).Scan(&origin)
	return sl, origin
}

// After an import let its download go because another app took the files,
// the movie is wanted with no download: the rescan links the copy that app
// placed, so the movie is not searched again.
func TestRescanLinksMovieCopyOfWantedMovie(t *testing.T) {
	svc, s, id, rel := movieWithFoundCopy(t, "")
	if _, err := svc.Rescan(context.Background(), domain.System); err != nil {
		t.Fatal(err)
	}
	sl, origin := linkedCopy(t, s, id, rel)
	if sl.FileID == nil || sl.State != mediafiles.SlotInLibrary || origin != "adopted" {
		t.Fatalf("slot %+v, origin %q", sl, origin)
	}
}

// A movie stuck from an older version on a download that is no longer
// running is linked too, and the stale download is let go.
func TestRescanLinksMovieOfFinishedDownload(t *testing.T) {
	svc, s, id, rel := movieWithFoundCopy(t, "finished")
	if _, err := svc.Rescan(context.Background(), domain.System); err != nil {
		t.Fatal(err)
	}
	sl, origin := linkedCopy(t, s, id, rel)
	if sl.FileID == nil || sl.ActiveDownloadID != nil || sl.State != mediafiles.SlotInLibrary || origin != "adopted" {
		t.Fatalf("slot %+v, origin %q", sl, origin)
	}
}

// A download still running keeps its movie: the found file stays a note.
func TestRescanLeavesMovieOfRunningDownload(t *testing.T) {
	svc, s, id, rel := movieWithFoundCopy(t, "downloading")
	if _, err := svc.Rescan(context.Background(), domain.System); err != nil {
		t.Fatal(err)
	}
	sl, origin := linkedCopy(t, s, id, rel)
	if sl.FileID != nil || sl.ActiveDownloadID == nil || sl.State != mediafiles.SlotNeedsADecision || origin != "found" {
		t.Fatalf("slot %+v, origin %q", sl, origin)
	}
}
