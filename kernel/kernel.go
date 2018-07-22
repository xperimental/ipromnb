package kernel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/prometheus/common/model"
	"github.com/xperimental/ipromnb/scaffold"
)

type Kernel struct {
	serverURL string
	client    *http.Client
	execution int
}

// New creates a new Prometheus kernel.
func New(server string) *Kernel {
	return &Kernel{
		serverURL: server,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (k *Kernel) HandleKernelInfo() scaffold.KernelInfo {
	return scaffold.KernelInfo{
		ProtocolVersion:       "5.2",
		Implementation:        "PrometheusKernel",
		ImplementationVersion: "0.0.1",
		LanguageInfo: scaffold.KernelLanguageInfo{
			Name: "prometheus",
		},
		Banner: "prometheus banner",
	}
}

func (k *Kernel) HandleExecuteRequest(ctx context.Context, req *scaffold.ExecuteRequest,
	stream func(name, text string),
	displayData func(data *scaffold.DisplayData, update bool)) *scaffold.ExecuteResult {

	k.execution++

	if strings.HasPrefix(req.Code, "server=") {
		k.serverURL = strings.TrimPrefix(req.Code, "server=")
		stream("stdout", fmt.Sprintf("Server set to %q.", k.serverURL))
		return &scaffold.ExecuteResult{
			Status:         "ok",
			ExecutionCount: k.execution,
		}
	}

	result, err := k.handleQuery(req.Code)
	if err != nil {
		stream("stderr", fmt.Sprintf("Error executing query: %s", err))
		return &scaffold.ExecuteResult{
			Status: "error",
		}
	}

	displayData(&scaffold.DisplayData{
		Data: map[string]interface{}{
			"text/html": result,
		},
	}, false)

	return &scaffold.ExecuteResult{
		Status:         "ok",
		ExecutionCount: k.execution,
	}
}

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

func (k *Kernel) handleQuery(query string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", k.serverURL, url.QueryEscape(query))
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

func (k *Kernel) HandleComplete(req *scaffold.CompleteRequest) *scaffold.CompleteReply {
	return &scaffold.CompleteReply{
		Status:      "ok",
		Matches:     []string{},
		CursorStart: req.CursorPos,
		CursorEnd:   req.CursorPos,
	}
}

func (k *Kernel) HandleInspect(req *scaffold.InspectRequest) *scaffold.InspectReply {
	return &scaffold.InspectReply{
		Status: "ok",
		Found:  false,
	}
}

func (k *Kernel) HandleIsComplete(req *scaffold.IsCompleteRequest) *scaffold.IsCompleteReply {
	return nil
}
