package fsx

//spellchecker:words path filepath testing
import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func makePaths(t *testing.T) (paths struct {
	Dir        string
	File       string
	Missing    string
	LinkToFile string
	LinkToDir  string
	BrokenLink string
}) {
	base := t.TempDir()

	paths.Dir = filepath.Join(base, "dir")
	paths.File = filepath.Join(base, "file")
	paths.Missing = filepath.Join(base, "missing")
	paths.LinkToDir = filepath.Join(base, "dirlink")
	paths.LinkToFile = filepath.Join(base, "filelink")
	paths.BrokenLink = filepath.Join(base, "brokenlink")

	if err := os.Mkdir(paths.Dir, fs.ModeDir); err != nil {
		panic(err)
	}

	if err := os.WriteFile(paths.File, nil, fs.ModePerm); err != nil {
		panic(err)
	}

	if err := os.Symlink(paths.File, paths.LinkToFile); err != nil {
		panic(err)
	}

	if err := os.Symlink(paths.Dir, paths.LinkToDir); err != nil {
		panic(err)
	}

	if err := os.Symlink(paths.Missing, paths.BrokenLink); err != nil {
		panic(err)
	}
	return
}

func TestExists(t *testing.T) {
	paths := makePaths(t)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"file", args{paths.File}, true, false},
		{"directory", args{paths.Dir}, true, false},
		{"missing", args{paths.Missing}, false, false},
		{"link to file", args{paths.LinkToFile}, true, false},
		{"link to directory", args{paths.LinkToDir}, true, false},
		{"broken link", args{paths.BrokenLink}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Exists(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDirectory(t *testing.T) {
	paths := makePaths(t)

	type args struct {
		path        string
		followLinks bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"file, follow=true", args{paths.File, true}, false, false},
		{"file, follow=false", args{paths.File, false}, false, false},
		{"directory, follow=true", args{paths.Dir, true}, true, false},
		{"directory, follow=false", args{paths.Dir, false}, true, false},
		{"missing, follow=true", args{paths.Missing, true}, false, false},
		{"missing, follow=false", args{paths.Missing, false}, false, false},
		{"link to file, follow=true", args{paths.LinkToFile, true}, false, false},
		{"link to file, follow=false", args{paths.LinkToFile, false}, false, false},
		{"link to directory, follow=true", args{paths.LinkToDir, true}, true, false},
		{"link to directory, follow=false", args{paths.LinkToDir, false}, false, false},
		{"broken link, follow=true", args{paths.BrokenLink, true}, false, false},
		{"broken link, follow=false", args{paths.BrokenLink, false}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsDirectory(tt.args.path, tt.args.followLinks)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRegular(t *testing.T) {
	paths := makePaths(t)
	type args struct {
		path        string
		followLinks bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"file, follow=true", args{paths.File, true}, true, false},
		{"file, follow=false", args{paths.File, false}, true, false},
		{"directory, follow=true", args{paths.Dir, true}, false, false},
		{"directory, follow=false", args{paths.Dir, false}, false, false},
		{"missing, follow=true", args{paths.Missing, true}, false, false},
		{"missing, follow=false", args{paths.Missing, false}, false, false},
		{"link to file, follow=true", args{paths.LinkToFile, true}, true, false},
		{"link to file, follow=false", args{paths.LinkToFile, false}, false, false},
		{"link to directory, follow=true", args{paths.LinkToDir, true}, false, false},
		{"link to directory, follow=false", args{paths.LinkToDir, false}, false, false},
		{"broken link, follow=true", args{paths.BrokenLink, true}, false, false},
		{"broken link, follow=false", args{paths.BrokenLink, false}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsRegular(tt.args.path, tt.args.followLinks)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsRegular() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsRegular() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsLink(t *testing.T) {
	paths := makePaths(t)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"file", args{paths.File}, false, false},
		{"directory", args{paths.Dir}, false, false},
		{"missing", args{paths.Missing}, false, false},
		{"link to file", args{paths.LinkToFile}, true, false},
		{"link to directory", args{paths.LinkToDir}, true, false},
		{"broken link", args{paths.BrokenLink}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsLink(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsLink() = %v, want %v", got, tt.want)
			}
		})
	}
}
