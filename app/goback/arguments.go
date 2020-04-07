package main

import (
	"flag"
)

// Arguments contains the values that are set from the command line
type Arguments struct {
	Source      string // Directory to backup - only one folder to make it simple.
	Target      string // Directory in which to create the timestamped folder and store the metadata
	OutputLevel int    // What detail to log to stdout
	Type        string // Backup type - translates to timestamp
	// ChangeDetection string
	NoProgress bool // Whether not to output progress information to StdOut
}

func (args *Arguments) fill() *Exit {
	flag.IntVar(&args.OutputLevel, "level", OutputLevelDefault, "Outputlevel: Debug = 1, Info = 2, Warning = 3, Error = 4")
	flag.StringVar(&args.Type, "type", "", "How often to create a new incremental backup directory: hourly, daily, monthly, yearly")
	flag.StringVar(&args.Source, "source", "", "The directory to backup")
	flag.BoolVar(&args.NoProgress, "no-progress", false, "Whether to suppress progress output to standard output")

	// Not supported yet
	// flag.StringVar(&args.ChangeDetection, "change", ChangeDetectionModificationAndSize, "Which type of change detection to use: modsize")

	showHelp := false
	flag.BoolVar(&showHelp, "help", false, "Show this help")

	flag.Parse()

	if !showHelp {
		arguments := flag.Args()

		if len(arguments) == 1 {
			args.Target = arguments[0]
		} else {
			Log.F(OutputLevelError, "The target directory must be the last argument")
			showHelp = true
		}
	}

	if showHelp {
		return &Exit{
			Code:     ExitCodeOk,
			Message:  "",
			ShowHelp: true,
		}
	}

	// Outputlevel and NoProgress are the only arguments that are not stored
	Log.Level = args.OutputLevel
	Log.NoProgress = args.NoProgress

	return nil
}
