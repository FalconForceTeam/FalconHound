package output_processor

import (
	"bufio"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type HTMLOutputConfig struct {
	Path             string
	QueryName        string
	QueryEventID     string
	QueryDescription string
	QueryDate        string
}

type HTMLOutputProcessor struct {
	*OutputProcessor
	Config HTMLOutputConfig
}

// HTML does not require batching, will write all output in one go
func (m *HTMLOutputProcessor) BatchSize() int {
	return 0
}

func (m *HTMLOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	err := WriteHTML(QueryResults, m.Config.Path)
	return err
}

func WriteHTML(results internal.QueryResults, path string) error {
	path = strings.Replace(path, "{{date}}", time.Now().Format("2006-01-02"), 2)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed creating directories: %w", err)
	}

	htmlFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed creating file: %w", err)
	}
	defer htmlFile.Close()

	writer := bufio.NewWriter(htmlFile)

	// Convert results to JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal results to JSON: %w", err)
	}

	// Write the HTML header
	htmlHeader := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Query Results</title>
    <script src="https://unpkg.com/react@17/umd/react.production.min.js"></script>
    <script src="https://unpkg.com/react-dom@17/umd/react-dom.production.min.js"></script>
    <script src="https://unpkg.com/babel-standalone@6/babel.min.js"></script>
    <script src="https://unpkg.com/axios/dist/axios.min.js"></script>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
    <script src="https://unpkg.com/ag-grid-community@32.2.1/dist/ag-grid-community.min.noStyle.js"></script>
    <link rel="stylesheet" href="https://unpkg.com/@ag-grid-community/styles@32.2.1/ag-grid.css">
    <link rel="stylesheet" href="https://unpkg.com/@ag-grid-community/styles@32.2.1/ag-theme-quartz.css">
</head>
<body>
<div id="root"></div>
<script type="text/babel">
    function QueryResults() {
        const data = ` + "`" + string(jsonData) + "`" + `;

        React.useEffect(() => {
            const gridOptions = {
                columnDefs: [
`
	_, err = writer.WriteString(htmlHeader)
	if err != nil {
		return fmt.Errorf("failed writing HTML header: %w", err)
	}

	// Write the column definitions
	if len(results) > 0 {
		for key := range results[0] {
			columnDef := fmt.Sprintf(`{ headerName: "%s", field: "%s" },`, key, key)
			_, err = writer.WriteString(columnDef)
			if err != nil {
				return fmt.Errorf("failed writing column definitions: %w", err)
			}
		}
	}

	// Write the row data
	htmlData := `
                ],
                rowData: JSON.parse(data),
                domLayout: 'autoHeight',
                pagination: true,
                paginationPageSize: 50,
                paginationPageSizeSelector: [50, 100, 250],
                defaultColDef: {
                    sortable: true,
                    filter: true,
                    resizable: true,
                },
				autoSizeStrategy: {
                        type: 'fitGridWidth',
                        defaultMinWidth: 200,
				        viewportMinWidth: 200,},
            };

            const eGridDiv = document.querySelector('#myGrid');
            new agGrid.Grid(eGridDiv, gridOptions);
        }, []);

        return (
            <div id="myGrid" className="ag-theme-quartz-dark" style={{ width: '100%', height: '500px' }}></div>
        );
    }

    ReactDOM.render(<QueryResults />, document.getElementById('root'));
</script>
</body>
</html>
`
	_, err = writer.WriteString(htmlData)
	if err != nil {
		return fmt.Errorf("failed writing HTML footer: %w", err)
	}

	writer.Flush()
	htmlFile.Close()

	return nil
}
