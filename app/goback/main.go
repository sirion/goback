// IDEA: Choose comparison method: hash, change timestamp + file size
// IDEA: Output progress: Add Additional Information, like MB/s
// IDEA: Option to display verify configuration for subsequent backups
package main

// Version should be increased for every release
const Version = "0.1"

func main() {
	// Fill Arguments structure from the command line
	args := &Arguments{}
	PerformExit(args.fill())

	backup := &Backup{}
	PerformExit(backup.loadConfiguration(args))
	PerformExit(backup.hash())
	PerformExit(backup.create())
}
