package kernel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/common/model"
	"github.com/xperimental/ipromnb/scaffold"
)

type Kernel struct {
	Options   Options
	client    *http.Client
	execution int
	queries   []string
}

// New creates a new Prometheus kernel.
func New(server string) *Kernel {
	return &Kernel{
		Options: Options{
			Server:    server,
			TimeStart: time.Now().Add(-24 * time.Hour),
			TimeEnd:   time.Now(),
			NowFunc:   time.Now,
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		execution: 0,
		queries:   []string{},
	}
}

func (k *Kernel) HandleKernelInfo() scaffold.KernelInfo {
	return scaffold.KernelInfo{
		ProtocolVersion:       "5.2",
		Implementation:        "ipromnb",
		ImplementationVersion: "0.0.1",
		LanguageInfo: scaffold.KernelLanguageInfo{
			Name: "prometheus",
		},
		Banner: "prometheus banner",
		HelpLinks: []scaffold.HelpLink{
			{
				Text: "PromQL Basics",
				URL:  "https://prometheus.io/docs/prometheus/latest/querying/basics/",
			},
			{
				Text: "PromQL Operators",
				URL:  "https://prometheus.io/docs/prometheus/latest/querying/operators/",
			},
			{
				Text: "PromQL Functions",
				URL:  "https://prometheus.io/docs/prometheus/latest/querying/functions/",
			},
		},
	}
}

var graphRegex = regexp.MustCompile(`^graph(0?)\((.+)\)$`)

func (k *Kernel) HandleExecuteRequest(ctx context.Context, req *scaffold.ExecuteRequest,
	stream func(name, text string), displayData scaffold.DisplayFunc) *scaffold.ExecuteResult {

	k.execution++

	if strings.HasPrefix(req.Code, "@") {
		if err := k.handleOptions(req.Code); err != nil {
			stream("stderr", fmt.Sprintf("Error setting options: %s", err))
			return &scaffold.ExecuteResult{
				Status: "error",
			}
		}

		stream("stdout", k.Options.Pretty())
		return &scaffold.ExecuteResult{
			Status:         "ok",
			ExecutionCount: k.execution,
		}
	}

	query, err := k.handleQuery(ctx, k.execution, req.Code, stream, displayData)
	k.queries = append(k.queries, query)

	if err != nil {
		stream("stderr", fmt.Sprintf("Error executing query: %s", err))
		return &scaffold.ExecuteResult{
			Status: "error",
		}
	}

	return &scaffold.ExecuteResult{
		Status:         "ok",
		ExecutionCount: k.execution,
	}
}

func lastIdentifier(input string, pos int) (string, int, int) {
	input = input[:pos]
	index := strings.LastIndexAny(input, ` +-*/{}[]()`)

	if index == -1 {
		return input, 0, len(input)
	}

	return input[index+1:], index + 1, len(input)
}

func (k *Kernel) HandleComplete(req *scaffold.CompleteRequest) *scaffold.CompleteReply {
	identifier, start, end := lastIdentifier(req.Code, req.CursorPos)

	api, err := k.getAPI()
	if err != nil {
		return &scaffold.CompleteReply{
			Status: "error",
		}
	}

	sets, err := api.Series(context.Background(), []string{fmt.Sprintf(`{__name__=~"%s.*"}`, identifier)}, k.Options.TimeStart, k.Options.TimeEnd)
	if err != nil {
		log.Printf("Error getting series: %s", err)
		return &scaffold.CompleteReply{
			Status: "error",
		}
	}

	metrics := map[string]bool{}
	for _, set := range sets {
		name, ok := set[model.MetricNameLabel]
		if ok {
			metrics[string(name)] = true
		}
	}

	matches := []string{}
	for k := range metrics {
		matches = append(matches, k)
	}
	sort.Strings(matches)

	return &scaffold.CompleteReply{
		Status:      "ok",
		Matches:     matches,
		CursorStart: start,
		CursorEnd:   end,
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
