package cli

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
)

func NewSpinner(msg string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 90*time.Millisecond)
	s.Suffix = " " + msg
	s.Writer = os.Stdout
	return s
}
