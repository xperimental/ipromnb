package kernel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"net/http"
	"net/url"
	"time"

	"github.com/prometheus/common/model"
)

type queryResult struct {
	Status string    `json:"status"`
	Data   queryData `json:"data"`
}

type queryData struct {
	Type   string        `json:"resultType"`
	Result []metricValue `json:"result"`
}

type metricValue struct {
	Labels model.Metric     `json:"metric"`
	Value  model.SamplePair `json:"value"`
}

func (k *Kernel) handleInstantQuery(query string, instant time.Time) (string, error) {
	q := url.QueryEscape(query)
	time := url.QueryEscape(instant.UTC().Format(time.RFC3339))
	url := fmt.Sprintf("%s/api/v1/query?query=%s&time=%s", k.Options.Server, q, time)
	res, err := k.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("http error: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-OK status code: %d", res.StatusCode)
	}

	var result queryResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error parsing result: %s", err)
	}

	output := &bytes.Buffer{}
	fmt.Fprintln(output, "<table><thead><tr><th>Metric</th><th>Value</th></thead><tbody>")
	for _, m := range result.Data.Result {
		fmt.Fprintln(output, fmt.Sprintf("<tr><td>%s</td><td>%f</td>", model.Metric(m.Labels), m.Value.Value))
	}
	fmt.Fprintf(output, "</tbody></table>")

	return output.String(), nil
}

func (k *Kernel) handleRangeQuery(query string, start, end time.Time) (*image.RGBA, error) {
	return nil, errors.New("not implemented")
}
