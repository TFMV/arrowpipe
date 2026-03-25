package cli

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCmd(t *testing.T) {
	dir, err := ioutil.TempDir("", "arrowpipe-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	rootCmd.SetArgs([]string{"init", dir})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	_, err = os.Stat(dir)
	assert.NoError(t, err)
}
