package kernel

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/gonum/plot/palette/brewer"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/xperimental/ipromnb/scaffold"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

var (
	errNoMetrics = errors.New("no matching metrics")
)

func (k *Kernel) handleQuery(ctx context.Context, count int, code string,
	stream func(name, text string), displayData scaffold.DisplayFunc) (string, error) {

	if match := graphRegex.FindStringSubmatch(code); match != nil {
		zero := match[1] == "0"
		query := match[2]

		result, err := k.handleRangeQuery(ctx, query, k.Options.TimeStart, k.Options.TimeEnd, zero)
		if err != nil {
			return "", err
		}

		displayData(&scaffold.DisplayData{
			Data: map[string]interface{}{
				"image/png": result,
			},
		}, false)

		return query, nil
	}

	result, err := k.handleInstantQuery(ctx, code, k.Options.TimeEnd)
	if err != nil {
		return "", err
	}

	displayData(&scaffold.DisplayData{
		Data: map[string]interface{}{
			"text/html": result,
		},
	}, false)

	return code, nil
}

func (k *Kernel) getAPI() (promv1.API, error) {
	if k.Options.Server == "" {
		return nil, fmt.Errorf("no server set. set one using @server=<url>")
	}

	client, err := api.NewClient(api.Config{
		Address: k.Options.Server,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %s", err)
	}

	return promv1.NewAPI(client), nil
}

func (k *Kernel) handleInstantQuery(ctx context.Context, query string, instant time.Time) (string, error) {
	api, err := k.getAPI()
	if err != nil {
		return "", err
	}

	value, err := api.Query(ctx, query, instant)
	if err != nil {
		return "", fmt.Errorf("query failed: %s", err)
	}

	result, ok := value.(model.Vector)
	if !ok {
		return "", fmt.Errorf("can not convert to vector: %t", value)
	}

	if len(result) == 0 {
		return "", errNoMetrics
	}

	output := &bytes.Buffer{}
	fmt.Fprintln(output, "<table><thead><tr><th>Metric</th><th>Value</th></thead><tbody>")
	for _, m := range result {
		fmt.Fprintln(output, fmt.Sprintf("<tr><td>%s</td><td>%f</td>", m.Metric, m.Value))
	}
	fmt.Fprintf(output, "</tbody></table>")

	return output.String(), nil
}

const (
	imageWidth  = 640
	imageHeight = 480
)

// Only show important part of metric name
var labelText = regexp.MustCompile("\\{(.*)\\}")

func (k *Kernel) handleRangeQuery(ctx context.Context, query string, start, end time.Time, zero bool) ([]byte, error) {
	duration := k.Options.TimeEnd.Sub(k.Options.TimeStart)
	rng := promv1.Range{
		Start: k.Options.TimeStart,
		End:   k.Options.TimeEnd,
		Step:  duration / 320,
	}

	api, err := k.getAPI()
	if err != nil {
		return nil, err
	}

	value, err := api.QueryRange(ctx, query, rng)
	if err != nil {
		return nil, fmt.Errorf("query failed: %s", err)
	}

	metrics, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("failed to convert to matrix: %t", value)
	}

	if len(metrics) == 0 {
		return nil, errNoMetrics
	}

	return plotResult(metrics, zero)
}

func plotResult(metrics model.Matrix, zero bool) ([]byte, error) {
	p, err := plot.New()
	if err != nil {
		return nil, fmt.Errorf("error creating plotter: %s", err)
	}

	textFont, err := vg.MakeFont("Helvetica", 3*vg.Millimeter)
	if err != nil {
		return nil, fmt.Errorf("failed to load font: %v", err)
	}

	p.X.Tick.Marker = plot.TimeTicks{Format: "15:04"}
	p.X.Tick.Label.Font = textFont
	p.Y.Tick.Label.Font = textFont
	p.Legend.Font = textFont
	p.Legend.YOffs = 10 * vg.Millimeter

	if zero {
		p.Y.Min = 0
	}

	// Color palette for drawing lines
	paletteSize := 8
	palette, err := brewer.GetPalette(brewer.TypeAny, "Dark2", paletteSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get color palette: %v", err)
	}
	colors := palette.Colors()

	for s, sample := range metrics {
		data := make(plotter.XYs, len(sample.Values))
		for i, v := range sample.Values {
			data[i].X = float64(v.Timestamp.Unix())
			f, err := strconv.ParseFloat(v.Value.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("sample value not float: %s", v.Value.String())
			}
			data[i].Y = f
		}

		l, err := plotter.NewLine(data)
		if err != nil {
			return nil, fmt.Errorf("failed to create line: %v", err)
		}
		l.LineStyle.Width = vg.Points(1)
		l.LineStyle.Color = colors[s%paletteSize]

		p.Add(l)
		if len(metrics) > 1 {
			m := labelText.FindStringSubmatch(sample.Metric.String())
			if m != nil {
				p.Legend.Add(m[1], l)
			}
		}
	}

	c, err := draw.NewFormattedCanvas(imageWidth, imageHeight, "png")
	if err != nil {
		return nil, fmt.Errorf("error creating canvas: %s", err)
	}

	p.Draw(draw.New(c))

	buf := &bytes.Buffer{}
	if _, err := c.WriteTo(buf); err != nil {
		return nil, fmt.Errorf("error writing image: %s", err)
	}

	return buf.Bytes(), nil
}
