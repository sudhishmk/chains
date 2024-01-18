/*
Copyright 2019 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pipelinerunmetrics

import (
	"context"
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
)

var (
	sgCount = stats.Float64("signed_total",
		"Total number of signed pipelineruns",
		stats.UnitDimensionless)

	sgCountView *view.View

	plCount = stats.Float64("payload_stored_total",
		"Total number of stored payloads for pipelineruns",
		stats.UnitDimensionless)

	plCountView *view.View
)

// Recorder holds keys for Tekton metrics
type Recorder struct {
	mutex           sync.Mutex
	initialized     bool
	ReportingPeriod time.Duration
}

// We cannot register the view multiple times, so NewRecorder lazily
// initializes this singleton and returns the same recorder across any
// subsequent invocations.
var (
	once           sync.Once
	r              *Recorder
	errRegistering error
)

// NewRecorder creates a new metrics recorder instance
// to log the PipelineRun related metrics
func NewRecorder(ctx context.Context) (*Recorder, error) {
	logger := logging.FromContext(ctx)
	once.Do(func() {
		r = &Recorder{
			initialized: true,

			// Default to 30s intervals.
			ReportingPeriod: 30 * time.Second,
		}
		errRegistering = viewRegister()
		logger.Debugf("View Register initiated ", r.initialized)
		if errRegistering != nil {
			r.initialized = false
			logger.Errorf("View Register Failed ", r.initialized)
			return
		}
	})

	return r, errRegistering
}

func viewRegister() error {
	//r.mutex.Lock()
	//defer r.mutex.Unlock()

	sgCountView = &view.View{
		Description: sgCount.Description(),
		Measure:     sgCount,
		Aggregation: view.Count(),
	}
	plCountView = &view.View{
		Description: plCount.Description(),
		Measure:     plCount,
		Aggregation: view.Count(),
	}

	return view.Register(
		sgCountView,
		plCountView,
	)
}

func viewUnregister() {
	view.Unregister(sgCountView, plCountView)
}

func (r *Recorder) RecordSignedCountMetrics(ctx context.Context, count int) {
	logger := logging.FromContext(ctx)
	logger.Debugf("Recording signed count metrics for context ", ctx, count)

	if !r.initialized {
		logger.Errorf("ignoring the metrics recording as recorder not initialized ")
	}
	count++
	r.countMetrics(ctx, float64(count), sgCount)
}

func (r *Recorder) RecordPayloadCountMetrics(ctx context.Context, count int) {
	logger := logging.FromContext(ctx)
	logger.Debugf("Recording payload count metrics for context ", ctx, count)

	if !r.initialized {
		logger.Errorf("ignoring the metrics recording as recorder not initialized ")
	}
	count++
	r.countMetrics(ctx, float64(count), plCount)
}

func (r *Recorder) countMetrics(ctx context.Context, count float64, measure *stats.Float64Measure) {
	logger := logging.FromContext(ctx)
	logger.Debugf("Recording metrics for context ", ctx, count, measure)
	if !r.initialized {
		logger.Errorf("ignoring the metrics recording for %s, failed to initialize the metrics recorder", measure.Description())
	}

	metrics.Record(ctx, measure.M(1))
}
