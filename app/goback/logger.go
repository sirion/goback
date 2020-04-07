package main

// #include <sys/ioctl.h>
// typedef struct winsize winsize;
// void myioctl(int i, unsigned long l, winsize * t){ioctl(i,l,t);}
import "C"
import (
	"fmt"
	"math"
	"os"
	"strings"
)

// TIOCGWINSZ value taken from c header file
const TIOCGWINSZ C.ulong = 0x5413

// Log is the global simple logger
var Log = Logger{}

// Level of logging to SdtErr
const (
	OutputLevelDefault = 3 // Default level: Log Warnings to StdErr

	OutputLevelDebug   = 1
	OutputLevelInfo    = 2
	OutputLevelWarning = 3
	OutputLevelError   = 4
	OutputLevelNone    = 5
)

var levelTags = map[int]string{
	OutputLevelDebug:   "[DEBUG] ",
	OutputLevelInfo:    "[INFO] ",
	OutputLevelWarning: "[WARN] ",
	OutputLevelError:   "[ERROR] ",
}

// Logger is a simple way of filering log output
type Logger struct {
	Level        int
	ProgressMax  float64
	ProgressStep float64
	NoProgress   bool
}

// F is like Fprint but prints to Stderr and only of the level is higher than the configured outputlevel
func (l *Logger) F(level int, format string, args ...interface{}) {
	if l.Level <= level {
		l.clearProgress()
		fmt.Fprintf(os.Stderr, levelTags[level]+format+"\n", args...)
		l.currentStep()
	}
}

// ProgressMessage prints a message inside the progress bar
func (l *Logger) ProgressMessage(message string) {
	if l.NoProgress {
		return
	}

	width := getTerminalWidth()
	if width < 10 {
		return
	}

	_, _ = fmt.Fprintf(os.Stdout, "\r[%s]", strings.Repeat(" ", int(width)-9))
	_, _ = fmt.Fprintf(os.Stdout, "\r[ %s", message)
}

// Step increases the progress by one and renders the current progress bar
func (l *Logger) Step() {
	if l.NoProgress {
		return
	}

	l.ProgressStep++

	l.currentStep()
}

func (l *Logger) currentStep() {
	if l.ProgressMax < 1 {
		return
	}

	done := l.ProgressStep / l.ProgressMax
	if done > 1 {
		return
	}

	width := getTerminalWidth()
	if width < 10 {
		return
	}

	contentWidth := width - 11

	bars := int(math.Ceil(contentWidth * done))
	spaces := int(math.Floor(contentWidth * (1 - done)))

	_, _ = fmt.Fprintf(os.Stdout, "\r[%s%s] %6.2f %%", strings.Repeat("=", bars), strings.Repeat(" ", spaces), done*100)

	if done == 1 {
		_, _ = fmt.Fprint(os.Stdout, "\n")
	}
}

func (l *Logger) clearProgress() {
	if l.NoProgress {
		return
	}
	width := getTerminalWidth()
	if width < 10 {
		return
	}

	_, _ = fmt.Fprintf(os.Stdout, "\r%s\r", strings.Repeat(" ", int(width)))
}

// TODO: Does this work on windows, mac, etc?
func getTerminalWidth() float64 {
	var ts C.winsize
	C.myioctl(0, TIOCGWINSZ, &ts)
	return float64(ts.ws_col)
}
