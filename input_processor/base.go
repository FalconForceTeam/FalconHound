package input_processor

import (
	"falconhound/internal"
	"falconhound/output_processor"
)

type InputProcessor struct {
	Type             string
	Enabled          bool
	Debug            bool
	Name             string
	ID               string
	SourcePlatform   string
	Credentials      internal.Credentials
	Query            string
	OutputProcessors []output_processor.OutputProcessorInterface
}

type InputProcessorInterface interface {
	ExecuteQuery() (internal.QueryResults, error)
	GetOutputProcessors() []output_processor.OutputProcessorInterface
	InputProcessorConfig() InputProcessor
}

func (m InputProcessor) GetOutputProcessors() []output_processor.OutputProcessorInterface {
	return m.OutputProcessors
}

func (m InputProcessor) InputProcessorConfig() InputProcessor {
	return m
}
