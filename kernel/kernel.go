package kernel

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

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
		Implementation:        "PrometheusKernel",
		ImplementationVersion: "0.0.1",
		LanguageInfo: scaffold.KernelLanguageInfo{
			Name: "prometheus",
		},
		Banner: "prometheus banner",
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
