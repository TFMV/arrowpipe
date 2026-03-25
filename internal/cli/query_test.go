package cli

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectCmd(t *testing.T) {
	dir, err := ioutil.TempDir("", "arrowpipe-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create dummy data
	datasetPath := filepath.Join(dir, "my.dataset")
	rootCmd.SetArgs([]string{"create-dummy-data", "--out", datasetPath, "--rows", "10"})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Select
	rootCmd.SetArgs([]string{"select", datasetPath, "--columns", "col1,col3"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
}
