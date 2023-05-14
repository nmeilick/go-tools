package tools

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ResolvePath resolves the given path. If it exist, it is returned. If it does not exist and does not contain
// any wildcard characters, os.ErrNotExist is returned. Otherwise, the result of filepath.Glob is returned.
// Unless the base of the glob pattern starts with a dot, entries stating with a dot are ignored.
func ResolvePath(path string) ([]string, error) {
	path = filepath.Clean(path)

	if _, err := os.Stat(path); err == nil {
		return []string{path}, nil
	} else if !strings.ContainsAny(path, "*?[") {
		return nil, os.ErrNotExist
	}

	matches, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}

	skipDot := !strings.HasPrefix(filepath.Base(path), ".")

	paths := []string{}
	for _, path = range matches {
		if _, err := os.Stat(path); err == nil {
			if skipDot && strings.HasPrefix(filepath.Base(path), ".") {
				continue
			}
			paths = append(paths, path)
		}
	}
	return paths, nil
}

// ResolveFiles resolves the given path to all existing files, see ResolvePath.
func ResolveFiles(path string) ([]string, error) {
	paths, err := ResolvePath(path)
	if err != nil {
		return nil, err
	}

	files := []string{}
	for _, path = range paths {
		if stat, err := os.Stat(path); err == nil && stat.Mode().IsRegular() {
			files = append(files, path)
		}
	}
	return files, nil
}

func SaveFileFunc(file string, f func(w io.Writer) error, perm os.FileMode) error {
	dir := filepath.Dir(file)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(file))
	if err != nil {
		// Return unless the error indicates that an intermediate directory may be missing
		if !os.IsNotExist(err) {
			return err
		}

		// Try to create the last directory in the path. Permissions are inferred from file read permission.
		dperm := perm | ((perm & 0444) >> 2)
		if err = os.Mkdir(dir, dperm); err != nil {
			return err
		}
		tmp, err = os.CreateTemp(dir, "."+filepath.Base(file))
		if err != nil {
			return err
		}
	}

	if err = f(tmp); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}

	if err = tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	if err = os.Rename(tmp.Name(), file); err != nil {
		os.Remove(tmp.Name())
	}
	return err
}

// SaveFile safely writes data to a file by writing it to a temporary file first before moving it over the
// destination file to ensure atomicity.
func SaveFile(file string, data []byte, perm os.FileMode) error {
	f := func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	}
	return SaveFileFunc(file, f, perm)
}

// SaveJSON safely writes JSON encoded data to a file by encoding the given value to a temporary file first
// before moving it over the destination file. This should ensure atomicity.
func SaveJSON(file string, v interface{}, indented bool, perm os.FileMode) error {
	f := func(w io.Writer) error {
		enc := json.NewEncoder(w)
		if indented {
			enc.SetIndent("", "  ")
		}
		return enc.Encode(v)
	}
	return SaveFileFunc(file, f, perm)
}

// LoadJSON decodes JSON read from the given file.
func LoadJSON(file string, v interface{}) error {
	h, err := os.Open(file)
	if err != nil {
		return err
	}
	defer h.Close()
	return json.NewDecoder(h).Decode(v)
}
