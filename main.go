package main

import (
	"errors"
	"falconhound/cmd"
	"falconhound/input_processor"
	"falconhound/internal"
	"falconhound/output_processor"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema"
	"gopkg.in/yaml.v2"
)

type Target struct {
	Name          string            `yaml:"Name"`
	Enabled       bool              `yaml:"Enabled"`
	Path          string            `yaml:"Path,omitempty"`
	Query         string            `yaml:"Query,omitempty"`
	Message       string            `yaml:"Message,omitempty"`
	Parameters    map[string]string `yaml:"Parameters,omitempty"`
	WatchlistName string            `yaml:"WatchlistName,omitempty"`
	Overwrite     bool              `yaml:"Overwrite,omitempty"`
	DisplayName   string            `yaml:"DisplayName,omitempty"`
	SearchKey     string            `yaml:"SearchKey,omitempty"`
	BHQuery       string            `yaml:"BHQuery,omitempty"`
	BatchSize     int               `yaml:"BatchSize,omitempty"`
}

type Query struct {
	Query          string   `yaml:"Query"`
	SourcePlatform string   `yaml:"SourcePlatform"`
	Name           string   `yaml:"Name"`
	Active         bool     `yaml:"Active"`
	Debug          bool     `yaml:"Debug"`
	Targets        []Target `yaml:"Targets"`
	Description    string   `yaml:"Description"`
	ID             string   `yaml:"ID"`
}

func main() {
	var runFlag bool
	flag.BoolVar(&runFlag, "go", false, "Run all actions in the actions directory")

	var helpFlag bool
	flag.BoolVar(&helpFlag, "help", false, "Print this help message")

	var actionsDir string
	flag.StringVar(&actionsDir, "actionsdir", "actions/", "Path to the actions directory")

	var keyvaultFlag bool
	flag.BoolVar(&keyvaultFlag, "keyvault", false, "Use the keyvault specified in the config for secrets")

	var configFile string
	flag.StringVar(&configFile, "config", "config.yml", "config file name")

	var ids string
	flag.StringVar(&ids, "ids", "", "comma separated list of action IDs to run")

	var debug bool
	flag.BoolVar((&debug), "debug", false, "Enable debug mode, for all executed actions")

	var actionlistFlag bool
	flag.BoolVar(&actionlistFlag, "actionlist", false, "Get a list of all enabled actions, use in combination with -go")

	flag.Parse()

	if helpFlag {
		printHelp()
		os.Exit(0)
	}

	internal.Banner()

	if runFlag {
		actionIdFilters := make([]string, 0)
		if ids != "" {
			// split ids on comma
			actionIdFilters = append(actionIdFilters, strings.Split(ids, ",")...)
		}
		run(actionsDir, configFile, keyvaultFlag, actionIdFilters, actionlistFlag, debug)
	} else {
		printHelp()
		os.Exit(0)
	}
}

func logError(errorLog *log.Logger, format string, v ...any) {
	log.Printf(cmd.Red+"[!] "+format+cmd.Reset, v...)
	errorLog.Printf("[!] "+format, v...)
}

func logInfo(format string, v ...any) {
	log.Printf(cmd.Blue+format+cmd.Reset, v...)
}

// Normalize Yaml data so it can be passed to the jsonschema validator
// This requires all map keys to be strings
func ensureStringKeys(val interface{}) (interface{}, error) {
	switch val := val.(type) {
	case map[interface{}]interface{}:
		// For a map convert all keys to strings
		// and apply recursively to all values
		m := make(map[string]interface{})
		for k, v := range val {
			k, ok := k.(string)
			if !ok {
				return nil, errors.New("found non-string key")
			}
			new_v, err := ensureStringKeys(v)
			if err != nil {
				return nil, err
			}
			m[k] = new_v
		}
		return m, nil
	case []interface{}:
		// For a list apply the normalization recursively to the
		// items in the list since these can be maps as well
		var err error
		var l = make([]interface{}, len(val))
		for i, v := range val {
			l[i], err = ensureStringKeys(v)
			if err != nil {
				return nil, err
			}
		}
		return l, nil
	default:
		return val, nil
	}
}

func validateQueryFile(yamlData []byte, actionsDir string) error {
	schemaPath := filepath.Join(actionsDir, "action_schema.json")
	schema, err := jsonschema.Compile(schemaPath)
	if err != nil {
		return err
	}
	var m interface{}
	err = yaml.Unmarshal([]byte(yamlData), &m)
	if err != nil {
		return err
	}
	m, err = ensureStringKeys(m)
	if err != nil {
		return err
	}

	if err := schema.ValidateInterface(m); err != nil {
		return err
	}
	return nil
}

func makeOutputProcessor(target Target, query Query, credentials internal.Credentials) (output_processor.OutputProcessorInterface, error) {
	baseOutput := output_processor.OutputProcessor{
		Enabled:     target.Enabled,
		Type:        target.Name,
		Credentials: credentials,
		Debug:       query.Debug,
	}
	switch target.Name {
	case "CSV":
		return &output_processor.CSVOutputProcessor{
			OutputProcessor: &baseOutput,
			Config: output_processor.CSVOutputConfig{
				Path: target.Path,
			},
		}, nil
	case "Sentinel":
		return &output_processor.SentinelOutputProcessor{
			OutputProcessor: &baseOutput,
			Config: output_processor.SentinelOutputConfig{
				QueryName:        query.Name,
				QueryDescription: query.Description,
				QueryEventID:     query.ID,
				BHQuery:          target.BHQuery,
			},
		}, nil
	case "Splunk":
		return &output_processor.SplunkOutputProcessor{
			OutputProcessor: &baseOutput,
			Config:          output_processor.SplunkOutputConfig{},
		}, nil
	case "Neo4j":
		return &output_processor.Neo4jOutputProcessor{
			OutputProcessor: &baseOutput,
			Config: output_processor.Neo4jOutputConfig{
				Parameters: target.Parameters,
				Query:      target.Query,
			},
		}, nil
	case "Watchlist":
		return &output_processor.WatchlistOutputProcessor{
			OutputProcessor: &baseOutput,
			Config: output_processor.WatchlistOutputConfig{
				WatchlistName: target.WatchlistName,
				DisplayName:   target.DisplayName,
				SearchKey:     target.SearchKey,
				Overwrite:     target.Overwrite,
			},
		}, nil
	case "BHSession":
		return &output_processor.BHSessionOutputProcessor{
			OutputProcessor: &baseOutput,
			Config: output_processor.BHSessionOutputConfig{
				BatchSize: target.BatchSize,
			},
		}, nil
	default:
		return nil, fmt.Errorf("Target %q not supported", target.Name)
	}
}

func makeInputProcessor(query Query, credentials internal.Credentials, outputs []output_processor.OutputProcessorInterface) (input_processor.InputProcessorInterface, error) {
	baseProcessor := input_processor.InputProcessor{
		Enabled:          query.Active,
		Debug:            query.Debug,
		Credentials:      credentials,
		Name:             query.Name,
		ID:               query.ID,
		SourcePlatform:   query.SourcePlatform,
		Query:            query.Query,
		OutputProcessors: outputs,
	}
	switch query.SourcePlatform {
	case "MDE":
		return &input_processor.MDEProcessor{
			InputProcessor: &baseProcessor,
			Config:         input_processor.MDEConfig{},
		}, nil
	case "Sentinel":
		return &input_processor.SentinelProcessor{
			InputProcessor: &baseProcessor,
			Config:         input_processor.SentinelConfig{},
		}, nil
	case "Neo4j":
		return &input_processor.Neo4jProcessor{
			InputProcessor: &baseProcessor,
			Config:         input_processor.Neo4jConfig{},
		}, nil
	case "BloodHound":
		return &input_processor.BHProcessor{
			InputProcessor: &baseProcessor,
			Config:         input_processor.BHConfig{},
		}, nil
	case "MSGraph":
		return &input_processor.MSGraphProcessor{
			InputProcessor: &baseProcessor,
			Config:         input_processor.MSGraphConfig{},
		}, nil
	default:
		return nil, fmt.Errorf("source platform %q not supported", query.SourcePlatform)
	}
}

func run(actionsDir string, configFile string, keyvaultFlag bool, actionIdFilters []string, actionlistFlag bool, debug bool) {
	// create error log
	fileError, err := openLogFile("./error.log")
	if err != nil {
		log.Fatal(err)
	}
	errorLog := log.New(fileError, "[error]", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)

	var files []string
	// define the function to be called for each file and directory visited
	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logError(errorLog, "Error visiting path %s: %v", path, err)
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}

	// recursively walk the directory and its subdirectories
	err = filepath.Walk(actionsDir, visit)
	if err != nil {
		logError(errorLog, "Error walking directory: %v", err)
		return
	}

	logInfo("[𓅃] Starting run")
	startTime := time.Now()

	ymlCount := 0
	for _, file := range files {
		if filepath.Ext(file) == ".yml" {
			ymlCount++
		}
	}
	log.Printf("[+] Found %d .yml files in %s", ymlCount, actionsDir)

	var queries []Query
	for _, filePath := range files {
		if filepath.Ext(filePath) == ".yml" {
			data, err := os.ReadFile(filePath)
			if err != nil {
				logError(errorLog, "failed to read file %s: %v", filePath, err)
				continue
			}

			var q Query
			if err := yaml.Unmarshal(data, &q); err != nil {
				logError(errorLog, "failed to unmarshal YAML in file %s: %v", filePath, err)
				continue
			}
			if err = validateQueryFile(data, actionsDir); err != nil {
				logError(errorLog, "failed to validate YAML in file %s: %v", filePath, err)
				continue
			}
			//count the active queries
			if !q.Active {
				continue
			}
			if len(actionIdFilters) > 0 {
				found := false
				for _, id := range actionIdFilters {
					if strings.EqualFold(strings.TrimSpace(id), strings.TrimSpace(q.ID)) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			queries = append(queries, q)
		}
	}

	if actionlistFlag {
		data := make([][]string, len(queries))
		for i, query := range queries {
			data[i] = []string{query.ID, query.Name, query.SourcePlatform}
		}
		headers := []string{"ID", "Name", "SourcePlatform"}
		// Print the table
		logInfo("[i] The following actions are available to run:")
		cmd.PrintTable(headers, data)

		// stop
		return
	}

	log.Printf("[+] Running %d active queries...", len(queries))

	globalCreds := cmd.GetCreds(configFile, keyvaultFlag)

	// output_processor.BHSubmit("test", globalCreds)

	var inputProcessors []input_processor.InputProcessorInterface = make([]input_processor.InputProcessorInterface, 0)

	for _, query := range queries {
		outputs := make([]output_processor.OutputProcessorInterface, 0)
		if debug {
			query.Debug = true
		}
		for _, target := range query.Targets {
			output, err := makeOutputProcessor(target, query, globalCreds)
			if err != nil {
				logError(errorLog, "Failed to create output processor for target %q: %v", target.Name, err)
				continue
			}
			outputs = append(outputs, output)
		}
		input, err := makeInputProcessor(query, globalCreds, outputs)
		if err != nil {
			logError(errorLog, "Failed to create input processor for target %q: %v", query.Name, err)
			continue
		}
		inputProcessors = append(inputProcessors, input)
	}

	for _, processor := range inputProcessors {
		inputProcessConfig := processor.InputProcessorConfig()
		logInfo("[Ξ] Running query %q (%s) in %s", inputProcessConfig.Name, inputProcessConfig.ID, inputProcessConfig.SourcePlatform)
		queryResults, err := processor.ExecuteQuery()
		if err != nil {
			logError(errorLog, "Error executing query: %v", err)
		}
		logInfo(" ↳ [>] Processing %v results..", len(queryResults))
		if inputProcessConfig.Debug {
			log.Printf("Results:")
			fmt.Printf("%#v\n", queryResults)
		}

		for _, outputProcessor := range processor.GetOutputProcessors() {
			outputProcessorConfig := outputProcessor.OutputProcessorConfig()
			if !outputProcessorConfig.Enabled {
				continue
			}
			logInfo("   ↳ [>] Writing to %s", outputProcessorConfig.Type)
			chunkSize := outputProcessor.BatchSize()
			if chunkSize < 0 {
				logError(errorLog, "Invalid batch size %d for %#v", chunkSize, outputProcessor)
				continue
			}
			// BatchSize() returns 0 if batching is not required and the output processor
			// can process all output at once
			if chunkSize == 0 {
				chunkSize = len(queryResults)
			}
			for i := 0; i < len(queryResults); i += chunkSize {
				end := i + chunkSize
				if end > len(queryResults) {
					end = len(queryResults)
				}
				err = outputProcessor.ProduceOutput(queryResults[i:end])
				if err != nil {
					logError(errorLog, "Error producing output for %#v: %#v", outputProcessor, err)
				}
			}
			err = outputProcessor.Finalize()
			if err != nil {
				logError(errorLog, "Error finalizing output for %#v: %w", outputProcessor, err)
			}
		}
	}
	logInfo("[=] All done ... finished in %.f seconds", time.Since(startTime).Seconds())
}

func printHelp() {
	fmt.Println("Usage: FalconHound -[options]")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}
