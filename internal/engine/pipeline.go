package engine

import (
	"context"

	"github.com/apache/arrow-go/v18/arrow"
)

// Stage is a function that processes a stream of records.
type Stage func(ctx context.Context, in <-chan arrow.Record) <-chan arrow.Record

// Pipeline represents a series of processing stages.
type Pipeline struct {
	stages []Stage
}

// NewPipeline creates a new pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{
		stages: make([]Stage, 0),
	}
}

// AddStage adds a processing stage to the pipeline.
func (p *Pipeline) AddStage(s Stage) {
	p.stages = append(p.stages, s)
}

// Execute runs the pipeline.
func (p *Pipeline) Execute(ctx context.Context, in <-chan arrow.Record) <-chan arrow.Record {
	var out <-chan arrow.Record = in
	for _, s := range p.stages {
		out = s(ctx, out)
	}
	return out
}
