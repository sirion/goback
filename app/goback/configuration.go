package main

import (
	"os"
	"path/filepath"
)

// Configuration contains everything that can be saved per backup
type Configuration struct {
	ChangeDetection   string `json:"change"`
	LastDirectoryName string `json:"last"`
	SourceDirectory   string `json:"source"`
	Format            string `json:"format"`
	targetDirectory   string
}

func (config *Configuration) fill(args *Arguments) (*Exit, bool) {

	// First try to load configuration from target
	exit, found := config.load(args.Target)
	if exit != nil {
		return exit, found
	}

	// Overwrite configuration from arguments
	exit = config.set(args)
	if exit != nil {
		return exit, found
	}

	return nil, found
}

// load reads the configuration from metadata file
func (config *Configuration) load(targetDirectory string) (*Exit, bool) {
	found, err := ReadJSON(filepath.Join(targetDirectory, ConfigurationFile), &config)
	if err != nil {
		return &Exit{
			Code:    ExitCodeConfigurationRead,
			Message: "Could not read configuration: " + err.Error(),
		}, false
	}

	config.targetDirectory = targetDirectory

	if !found {
		return nil, false
	}

	return nil, true
}

func (config *Configuration) save() *Exit {
	err := WriteJSON(filepath.Join(config.targetDirectory, ConfigurationFile), config)
	if err != nil {
		return &Exit{
			Code:    ExitCodeConfigurationWrite,
			Message: "Could not write configuration: " + err.Error(),
		}
	}

	return nil
}

func (config *Configuration) set(args *Arguments) *Exit {
	var err error

	showHelp := false

	// TODO: Overwrite current configuration with arguments
	// TODO: Write config file if something changed

	// Check if target exists
	// Check if target is a a directory
	// TODO: Check if target is writable
	if args.Target == "" {
		showHelp = true
		Log.F(OutputLevelError, "Please provide target directory as unnamed argument")
	} else {
		config.targetDirectory, err = filepath.Abs(args.Target)
		if err != nil {
			showHelp = true
			Log.F(OutputLevelError, "Target directory is not valid")
		}

		target, err := os.Stat(config.targetDirectory)
		if err != nil {
			showHelp = true
			Log.F(OutputLevelError, "Cannot access target directory %s: %s", config.targetDirectory, err.Error())
		} else if !target.IsDir() {
			showHelp = true
			Log.F(OutputLevelError, "Source is not a directory")
		}

	}

	// Check if source exists
	// Check if source is a directory
	if args.Source != "" {
		config.SourceDirectory, err = filepath.Abs(args.Source)
		if err != nil {
			showHelp = true
			Log.F(OutputLevelError, "Source directory is not valid")
		}
	}

	source, err := os.Stat(config.SourceDirectory)
	if err != nil {
		showHelp = true
		Log.F(OutputLevelError, "Cannot access source directory %s: %s", config.SourceDirectory, err.Error())
	} else if !source.IsDir() {
		showHelp = true
		Log.F(OutputLevelError, "Source is not a directory")
	}

	// Not supported yet:

	// switch args.ChangeDetection {
	// case ChangeDetectionModificationAndSize:
	// 	config.ChangeDetection = ChangeDetectionModificationAndSize
	// 	break

	// default:
	// 	Log.F(OutputLevelWarning, "Invalid change detection method, defaulting to %s", ChangeDetectionModificationAndSize)
	// 	config.ChangeDetection = ChangeDetectionModificationAndSize
	// 	break
	// }

	config.ChangeDetection = ChangeDetectionModificationAndSize

	if args.Type != "" {
		var ok bool
		config.Format, ok = Type2TimestampFormat[args.Type]
		if !ok {
			config.Format = ""
		}
	}

	if config.Format == "" {
		showHelp = true
		Log.F(OutputLevelError, "Backup type must be specified on first backup")
	}

	if showHelp {
		return &Exit{
			Code:     ExitCodeConfiguration,
			Message:  "",
			ShowHelp: true,
		}
	}

	return nil

}
