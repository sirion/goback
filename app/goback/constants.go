package main

// Timestamps relating to the backup types
const (
	TimestampFormatYear  = "2006"
	TimestampFormatMonth = "2006-01"
	TimestampFormatDay   = "2006-01-02"
	TimestampFormatHour  = "2006-01-02-15"
	// TimestampFormatMinute = "2006-01-02-15-04"
	// TimestampFormatSecond = "2006-01-02-15-04-05"
	TimestampFormatTest = "2006-01-02-15-04-05.999" // For testing so we can create new backups in short time

	// TimestampFormatHash is used for change detection when using "modsize"
	TimestampFormatHash = "20060102150405"
)

// Types of backups supported. This describes when a new folder will be created
const (
	BackupTypeHourly  = "hourly"
	BackupTypeDaily   = "daily"
	BackupTypeMonthly = "monthly"
	BackupTypeYearly  = "yearly"
	// Only for testing:
	BackupTypeTest = "test"
)

// Currently only "modsize" is available for the change detection
const (
	ChangeDetectionModificationAndSize = "modsize"
)

// ConfigurationFile is the name of the main metadata file in the backup directory
const ConfigurationFile = "config.goback"

// HashesExtension is the extension used for the files storing the hashes
const HashesExtension = "goback"

// Exit codes in case of an error
const (
	ExitCodeOk                 = 0
	ExitcodeNotCreated         = 1
	ExitcodeNoAccess           = 2
	ExitcodeCopyRead           = 3
	ExitcodeCopyCreate         = 4
	ExitcodeCopyWrite          = 5
	ExitcodeCopyClose          = 6
	ExitcodeCopyRename         = 7
	ExitcodeHashesMarshal      = 8
	ExitcodeHashesWrite        = 9
	ExitcodeReadDirectory      = 10
	ExitcodeNoReference        = 11
	ExitcodeWriteInfo          = 12
	ExitcodeCopyCreateDir      = 13
	ExitcodeNoType             = 14
	ExitCodeConfiguration      = 15
	ExitCodeConfigurationRead  = 16
	ExitCodeConfigurationWrite = 17
	ExitcodeCleanup            = 18

	ExitcodeOutput = 99
)

// Type2TimestampFormat is a mapping to set the timestamp using the backup type
var Type2TimestampFormat = map[string]string{
	BackupTypeHourly:  TimestampFormatHour,
	BackupTypeDaily:   TimestampFormatDay,
	BackupTypeMonthly: TimestampFormatMonth,
	BackupTypeYearly:  TimestampFormatYear,
	BackupTypeTest:    TimestampFormatTest,
}
