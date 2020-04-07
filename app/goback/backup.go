package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Backup is the main goback structure
type Backup struct {
	Initial       bool
	Configuration Configuration

	From string
	To   string
	Ref  string

	FromHashes map[string]string
	RefHashes  map[string]string
}

func (backup *Backup) loadConfiguration(args *Arguments) *Exit {
	// Initialize backup with Arguments
	exit, found := backup.Configuration.fill(args)
	if exit != nil {
		return exit
	}

	// The backup folder or no configration file found. Must be an initial backup
	backup.Initial = !found

	exit = backup.setup()
	if exit != nil {
		return exit
	}

	return nil
}

func (backup *Backup) setup() *Exit {
	config := backup.Configuration

	backup.From = config.SourceDirectory
	backup.To = filepath.Join(config.targetDirectory, time.Now().Format(config.Format))

	// Read metadata
	// Get name of the reference directory
	if config.LastDirectoryName != "" {
		backup.Ref = filepath.Join(config.targetDirectory, config.LastDirectoryName)
		if !DirectoryExists(backup.Ref) {
			return &Exit{
				Message: "Reference directory does not exist or cannot be accessed",
				Code:    ExitcodeNoReference,
			}
		}
	}

	if backup.Initial && config.Format == "" {
		return &Exit{
			Message: "Backup type must be provided for initial backup",
			Code:    ExitcodeNoType,
		}
	}

	// Create new backup directory
	err := os.Mkdir(backup.To, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return &Exit{
			Message: fmt.Sprintf("New backup directory could not be created - %s: %s", backup.To, err.Error()),
			Code:    ExitcodeNotCreated,
		}
	} else if os.IsExist(err) {
		Log.F(OutputLevelWarning, "Backup directory already exists. Resuming backup into %s", backup.To)
	}

	// Make sure all paths are normalized
	if backup.From != "" {
		backup.From, err = filepath.Abs(backup.From)
		if err != nil {
			return &Exit{
				Message: fmt.Sprintf("Could not normalize directory - %s: %s", backup.From, err.Error()),
				Code:    ExitcodeNotCreated,
			}
		}
	}

	if backup.To != "" {
		backup.To, err = filepath.Abs(backup.To)
		if err != nil {
			return &Exit{
				Message: fmt.Sprintf("Could not normalize directory - %s: %s", backup.To, err.Error()),
				Code:    ExitcodeNotCreated,
			}
		}
	}
	if backup.Ref != "" {
		backup.Ref, err = filepath.Abs(backup.Ref)
		if err != nil {
			return &Exit{
				Message: fmt.Sprintf("Could not normalize directory - %s: %s", backup.Ref, err.Error()),
				Code:    ExitcodeNotCreated,
			}
		}
	}

	return nil
}

func (backup *Backup) hash() *Exit {
	// Create hashes for backup data and save new hashes next to backup data
	var exit *Exit

	backup.FromHashes, exit = createHashes(backup.From, backup.To+"."+HashesExtension)
	if exit != nil {
		return exit
	}

	if !backup.Initial {
		// Read hashes for Reference
		backup.RefHashes, exit = getHashes(backup.Ref)
		if exit != nil {
			return exit
		}
	} else {
		backup.RefHashes = make(map[string]string)
	}

	return nil
}

func (backup *Backup) create() *Exit {

	if backup.Initial {
		Log.F(OutputLevelInfo, "Creating initial copy in %s", backup.To)
	}

	// TODO: Cleanup referenceName to prevent path traversal attacks

	// TODO: Check if new backup directory exists

	// Save last backup reference
	// TODO: Update Configuration.LastDirectoryName	and save config
	Log.F(OutputLevelDebug, "Saving backup info in configuration file")
	backup.Configuration.LastDirectoryName = filepath.Base(backup.To)
	exit := backup.Configuration.save()
	if exit != nil {
		return exit
	}

	// TODO: Go through list of files and compare to reference
	Log.F(OutputLevelInfo, "Backup of %d files...", len(backup.FromHashes))
	Log.ProgressMax = float64(len(backup.FromHashes))
	for filePath, hash := range backup.FromHashes {
		exit := backup.handleFile(filePath, hash)
		if exit != nil {
			return exit
		}
	}

	// TODO: Remove empty directories
	exit = CleanDirectory(backup.Ref)
	if exit != nil {
		return exit
	}

	return nil
}

func (backup *Backup) handleFile(filePath, hash string) *Exit {
	defer Log.Step()

	pathOri := filepath.Join(backup.From, filePath)
	pathNew := filepath.Join(backup.To, filePath)
	pathRef := filepath.Join(backup.Ref, filePath)

	_, err := os.Stat(pathNew)

	if err == nil {
		// If already exists in new backup directory, skip
		// TODO: Check for reference anyway?
		Log.F(OutputLevelInfo, "Skipping: %s", pathOri)
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		Log.F(OutputLevelError, "Could not access %s: %s", pathNew, err.Error())
		return nil
		// IDEA: OPtion to exit on error?
		// return &Exit{
		// 	Message: fmt.Sprintf("Could not access %s: %s", pathNew, err.Error()),
		// 	Code:    ExitcodeNoAccess,
		// }
	}

	if hash == backup.RefHashes[filePath] {
		// If same, move from reference to new backup directory
		Log.F(OutputLevelInfo, "Moving from last backup: %s", pathOri)
		exit := MoveFile(pathRef, pathNew)
		if exit != nil {
			return exit
		}

	} else {
		// TODO: If differs, copy source to new backup directory
		Log.F(OutputLevelInfo, "Copying: %s", pathOri)
		exit := CopyFile(pathOri, pathNew)
		if exit != nil {
			return exit
		}
	}

	return nil
}

///
///
/// Functions
///
///

func createHashes(directory, file string) (map[string]string, *Exit) {
	hashes := map[string]string{}

	Log.ProgressMessage(fmt.Sprintf("Creating hashes for directory %s...", directory))

	exit := hashDirectory(directory, "", hashes)
	if exit != nil {
		return nil, exit
	}

	hashData, err := json.Marshal(hashes)
	if err != nil {
		return nil, &Exit{
			Message: fmt.Sprintf("ERROR: Could not save hashes: %s", err.Error()),
			Code:    ExitcodeHashesMarshal,
		}
	}

	Log.F(OutputLevelDebug, "Saving hashes for %s in %s", directory, file)
	err = ioutil.WriteFile(file, hashData, os.ModePerm)
	if err != nil {
		return nil, &Exit{
			Message: fmt.Sprintf("ERROR: Could not save hashes in %s: %s", file, err.Error()),
			Code:    ExitcodeHashesWrite,
		}
	}

	return hashes, nil
}

func hashDirectory(dir, prefix string, hashes map[string]string) *Exit {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return &Exit{
			Message: fmt.Sprintf("ERROR: Could not read directory %s: %s", dir, err.Error()),
			Code:    ExitcodeReadDirectory,
		}
	}

	for _, file := range files {
		name := file.Name()
		if file.IsDir() {
			hashDirectory(filepath.Join(dir, name), prefix+name+"/", hashes)
		} else {
			// TODO: Actually calculate hash
			// ChangeDetectionModificationAndSize

			hashes[prefix+name] = file.ModTime().Format(TimestampFormatHash) + "|" + strconv.FormatInt(file.Size(), 10)
		}
	}

	return nil
}

func getHashes(dir string) (map[string]string, *Exit) {
	hashes := make(map[string]string)
	hashesFound := false
	hashFile := dir + "." + HashesExtension

	Log.F(OutputLevelDebug, "Reading hashes from %s", hashFile)
	hashData, err := ioutil.ReadFile(hashFile)
	if err != nil {
		Log.F(OutputLevelWarning, "Could not read hashes for %s: %s", hashFile, err.Error())
	} else {
		err = json.Unmarshal(hashData, &hashes)
		if err != nil {
			Log.F(OutputLevelWarning, "Could not read hashes from %s: %s", hashFile, err.Error())
		} else {
			hashesFound = true
		}
	}

	// If no hashes for reference cannot be found, create them
	if !hashesFound {
		var exit *Exit
		hashes, exit = createHashes(dir, hashFile)
		if exit != nil {
			return nil, exit
		}
	}

	return hashes, nil
}
