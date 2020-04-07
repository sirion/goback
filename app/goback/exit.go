package main

import (
	"flag"
	"fmt"
	"os"
)

// Exit describes the program termination
type Exit struct {
	Message  string
	Code     int
	ShowHelp bool
}

func (*Exit) showHelp() {
	echo("\n")
	echo("goback v%s\n", Version)
	echo("\n")
	echo("Usage of goback:\n\n")
	echo("  Initial backup:\n")
	echo("    goback [-type daily] [-change modsize] [-level 3] -source SOURCE TARGET \n")
	echo("\n")
	echo("  Subsequent backups:\n")
	echo("    goback [-type daily] [-level 3] [-source SOURCE] TARGET\n")
	echo("\n")
	echo("The arguments from the first backup will be saved inside the configuration (except level)\n")
	echo("\n")
	flag.PrintDefaults()
	os.Exit(0)
}

// PerformExit ends the application in case the given argument is an an exit.
func PerformExit(exit *Exit) {
	if exit != nil {
		if exit.Code != ExitCodeOk {
			Log.F(OutputLevelError, exit.Message)
		}
		if exit.ShowHelp {
			exit.showHelp()
		}
		os.Exit(exit.Code)
	}
}

func echo(format string, args ...interface{}) {
	_, err := fmt.Fprintf(flag.CommandLine.Output(), format, args...)
	if err != nil {
		os.Exit(ExitcodeOutput)
	}

}
