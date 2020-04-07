package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ReadJSON reads the given file and fills the given structure pointer, returns true ans second return in case the file is not found
func ReadJSON(path string, structure interface{}) (bool, error) {
	Log.F(OutputLevelDebug, "Reading from %s", path)

	data, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		Log.F(OutputLevelDebug, "File does not exist: %s", path)
		return false, nil
	} else if err != nil {
		return false, err
	}

	err = json.Unmarshal(data, structure)
	if err != nil {
		return false, err
	}

	return true, nil
}

// WriteJSON writes a JSON representation for the given structure into the given path
func WriteJSON(path string, structure interface{}) error {
	Log.F(OutputLevelDebug, "Writing to %s", path)

	data, err := json.Marshal(structure)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// DirectoryExists returns true if the given path exists and is a directory
func DirectoryExists(dir string) bool {
	info, err := os.Stat(dir)

	if err != nil {
		if !os.IsNotExist(err) {
			Log.F(OutputLevelError, "Error checking for directory %s: %s", dir, err.Error())
		}
		return false
	}

	if !info.IsDir() {
		Log.F(OutputLevelWarning, "Not a directory: %s ", dir)
		return false
	}
	return true
}

// MoveFile renames a file and creates directories if needed
func MoveFile(source, destination string) *Exit {
	destinationDir := filepath.Dir(destination)
	err := os.MkdirAll(destinationDir, os.ModePerm)
	if err != nil {
		return &Exit{
			Message: fmt.Sprintf("Creating folder %s: %s", destinationDir, err.Error()),
			Code:    ExitcodeCopyCreateDir,
		}
	}

	err = os.Rename(source, destination)
	if err != nil {
		return &Exit{
			Message: fmt.Sprintf("Moving file to %s: %s", destination, err.Error()),
			Code:    ExitcodeCopyCreateDir,
		}
	}

	return nil
}

// CopyFile creates a new file and directories if needed and copies the data from source to destination
func CopyFile(source, destination string) *Exit {
	in, err := os.Open(source)
	if err != nil {
		return &Exit{
			Message: fmt.Sprintf("Opening %s: %s", source, err.Error()),
			Code:    ExitcodeCopyRead,
		}
	}
	defer LogError(in.Close)

	destinationDir := filepath.Dir(destination)
	err = os.MkdirAll(destinationDir, os.ModePerm)
	if err != nil {
		return &Exit{
			Message: fmt.Sprintf("Creating target directory %s: %s", destinationDir, err.Error()),
			Code:    ExitcodeCopyCreateDir,
		}
	}

	pathTmp := destination + ".part"
	tmp, err := os.Create(pathTmp)
	if err != nil {
		return &Exit{
			Message: fmt.Sprintf("Creating temp file %s: %s", pathTmp, err.Error()),
			Code:    ExitcodeCopyCreate,
		}
	}

	_, err = io.Copy(tmp, in)
	if err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return &Exit{
			Message: fmt.Sprintf("Writing to %s: %s", pathTmp, err.Error()),
			Code:    ExitcodeCopyWrite,
		}
	}
	if err = tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return &Exit{
			Message: fmt.Sprintf("Closing %s: %s", pathTmp, err.Error()),
			Code:    ExitcodeCopyClose,
		}
	}

	err = os.Rename(pathTmp, destination)
	if err != nil {
		return &Exit{
			Message: fmt.Sprintf("Renaming %s: %s", pathTmp, err.Error()),
			Code:    ExitcodeCopyRename,
		}
	}
	return nil
}

// CleanDirectory deletes all empty folders recursively in the given directory
func CleanDirectory(directory string) *Exit {
	if directory == "" {
		return nil
	}

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return &Exit{
			Message: fmt.Sprintf("Cleaning directory %s: %s", directory, err.Error()),
			Code:    ExitcodeCleanup,
		}
	}

	hasContent := false
	for _, file := range files {
		path := filepath.Join(directory, file.Name())
		if file.IsDir() {
			exit := CleanDirectory(path)
			if exit != nil {
				return exit
			}
			_, err := os.Stat(path)
			if err != nil && os.IsNotExist(err) {
				// Directory has been deleted. No content for parent
			} else if err != nil {
				return &Exit{
					Message: fmt.Sprintf("Cleaning directory %s: %s", path, err.Error()),
					Code:    ExitcodeCleanup,
				}
			} else {
				hasContent = true
			}
		} else {
			hasContent = true
		}
	}

	if !hasContent {
		err := os.Remove(directory)
		if err != nil {
			return &Exit{
				Message: fmt.Sprintf("Cleaning directory, error deleting %s: %s", directory, err.Error()),
				Code:    ExitcodeCleanup,
			}
		}
	}

	return nil
}

// LogError is a helper function to avoid silencing errors when using defer. Use like this: "defer LogError(xxx.Close())"
func LogError(args ...func() error) {
	for _, errFn := range args {
		err := errFn()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}
