package kernel

import (
	"context"
	"sort"
	"strings"
)

func (k *Kernel) handleComplete(input string, cursorPos int) (matches []string, start, end int, err error) {
	identifier, start, end := lastIdentifier(input, cursorPos)

	api, err := k.getAPI()
	if err != nil {
		return nil, 0, 0, err
	}

	values, err := api.LabelValues(context.Background(), "__name__")
	if err != nil {
		return nil, 0, 0, err
	}

	metrics := map[string]bool{}
	for _, value := range values {
		valueStr := string(value)
		if strings.HasPrefix(valueStr, identifier) {
			matches = append(matches, valueStr)
		}
	}

	for k := range metrics {
		matches = append(matches, k)
	}
	sort.Strings(matches)

	return matches, start, end, nil
}
