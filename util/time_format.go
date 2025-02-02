package util

import (
	"fmt"
	"time"
)

func HumanReadableDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
}
