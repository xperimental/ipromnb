package kernel

import (
	"fmt"
	"strings"
	"time"
)

// Options holds the current options of the kernel.
// They can be set using commands.
type Options struct {
	Server    string
	TimeStart time.Time
	TimeEnd   time.Time
}

func (o Options) Pretty() string {
	start := o.TimeStart.UTC().Format(time.RFC3339)
	end := o.TimeEnd.UTC().Format(time.RFC3339)
	return fmt.Sprintf("Server: %s\n  Time: %s - %s", o.Server, start, end)
}

func (k *Kernel) handleOptions(input string) error {
	commands := strings.Split(input, "\n")
	for _, c := range commands {
		tokens := strings.SplitN(strings.TrimPrefix(c, "@"), "=", 2)
		if len(tokens) != 2 {
			return fmt.Errorf("not an assignment: %s", c)
		}

		key := strings.TrimSpace(strings.ToLower(tokens[0]))
		value := strings.TrimSpace(tokens[1])
		switch key {
		case "server":
			k.Options.Server = value
		case "timestart", "start":
			if err := setTime(&k.Options.TimeStart, value); err != nil {
				return err
			}
		case "timeend", "end":
			if err := setTime(&k.Options.TimeEnd, value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("not a valid option: %s", key)
		}
	}
	return nil
}

func setTime(v *time.Time, value string) error {
	if strings.ToLower(value) == "now" {
		*v = time.Now()
		return nil
	}

	time, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return fmt.Errorf("not a valid timestamp: %s", value)
	}

	*v = time
	return nil
}
