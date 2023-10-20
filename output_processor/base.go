package output_processor

import "falconhound/internal"

type OutputProcessor struct {
	Type        string
	Enabled     bool
	Credentials internal.Credentials
	Debug       bool
}

type OutputProcessorInterface interface {
	ProduceOutput(internal.QueryResults) error
	Finalize() error
	BatchSize() int
	OutputProcessorConfig() OutputProcessor
}

func (m *OutputProcessor) OutputProcessorConfig() OutputProcessor {
	return *m
}

// Can be used to perform any cleanup operations after inputs have been processed
func (m *OutputProcessor) Finalize() error {
	return nil
}
