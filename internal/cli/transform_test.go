package cli

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportExportCsv(t *testing.T) {
	dir, err := ioutil.TempDir("", "arrowpipe-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create dummy CSV
	csvPath := filepath.Join(dir, "input.csv")
	csvContent := `h1,h2
v1,v2
`
	err = ioutil.WriteFile(csvPath, []byte(csvContent), 0644)
	assert.NoError(t, err)

	// Import
	datasetPath := filepath.Join(dir, "my.dataset")
	rootCmd.SetArgs([]string{"import", "csv", csvPath, "--out", datasetPath})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Export
	exportedCsvPath := filepath.Join(dir, "output.csv")
	rootCmd.SetArgs([]string{"export", "csv", datasetPath, "--out", exportedCsvPath})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Compare
	inputContent, err := ioutil.ReadFile(csvPath)
	assert.NoError(t, err)
	exportedContent, err := ioutil.ReadFile(exportedCsvPath)
	assert.NoError(t, err)
	assert.Equal(t, inputContent, exportedContent)
}
