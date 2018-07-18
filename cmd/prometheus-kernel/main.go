package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/common/model"
	scaffold "github.com/yunabe/lgo/jupyter/gojupyterscaffold"
)

var (
	configFile string
	serverURL  string
)

func main() {
	flag.StringVar(&configFile, "connection-file", "", "Path to connection file.")
	flag.StringVar(&serverURL, "server-url", "", "Default Prometheus server.")
	flag.Parse()

	if configFile == "" {
		glog.Fatal("Need to provide a connection file.")
	}

	kernel := newPrometheusKernel(serverURL)

	server, err := scaffold.NewServer(configFile, kernel)
	if err != nil {
		glog.Fatalf("Error creating server: %s", err)
	}

	server.Loop()
}

type kernel struct {
	serverURL string
	client    *http.Client
	execution int
}

func newPrometheusKernel(server string) *kernel {
	return &kernel{
		serverURL: server,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (k *kernel) HandleKernelInfo() scaffold.KernelInfo {
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

func (k *kernel) HandleExecuteRequest(ctx context.Context, req *scaffold.ExecuteRequest,
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

func (k *kernel) handleQuery(query string) (string, error) {
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

func (k *kernel) HandleComplete(req *scaffold.CompleteRequest) *scaffold.CompleteReply {
	return &scaffold.CompleteReply{
		Status:      "ok",
		Matches:     []string{},
		CursorStart: req.CursorPos,
		CursorEnd:   req.CursorPos,
	}
}

func (k *kernel) HandleInspect(req *scaffold.InspectRequest) *scaffold.InspectReply {
	return &scaffold.InspectReply{
		Status: "ok",
		Found:  false,
	}
}

func (k *kernel) HandleIsComplete(req *scaffold.IsCompleteRequest) *scaffold.IsCompleteReply {
	return nil
}

func (k *kernel) HandleGoFmt(req *scaffold.GoFmtRequest) (*scaffold.GoFmtReply, error) {
	return nil, errors.New("not implemented")
}
