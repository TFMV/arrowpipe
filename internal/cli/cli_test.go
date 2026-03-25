package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
)

func TestVersionCmd(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})
	rootCmd.Execute()

	output, err := ioutil.ReadAll(buf)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(strings.TrimSpace(string(output)), "arrowpipe version"))
}

func createTestDataset(t *testing.T, dir string) {
	pool := memory.NewGoAllocator()
	schema := arrow.NewSchema(
		[]arrow.Field{{Name: "f1", Type: arrow.PrimitiveTypes.Int64}}, nil)

	file, err := os.Create(filepath.Join(dir, "data.arrow"))
	assert.NoError(t, err)
	defer file.Close()

	ds, err := storage.NewColumnarDataset(file, schema)
	assert.NoError(t, err)

	b := array.NewInt64Builder(pool)
	defer b.Release()
	b.AppendValues([]int64{1, 2, 3}, nil)
	arr := b.NewArray()
	defer arr.Release()
	rec := array.NewRecord(schema, []arrow.Array{arr}, 3)

	err = ds.Write(rec)
	assert.NoError(t, err)
	err = ds.Close()
	assert.NoError(t, err)
}

func withTempFile(t *testing.T, fn func(file *os.File)) {
	file, err := ioutil.TempFile("", "arrowpipe-test")
	assert.NoError(t, err)
	defer os.Remove(file.Name())
	defer file.Close()

	fn(file)
}

func TestSchemaCmd_Dataset(t *testing.T) {
	dir, err := ioutil.TempDir("", "arrowpipe-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	createTestDataset(t, dir)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"schema", dir})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	output, err := ioutil.ReadAll(buf)
	assert.NoError(t, err)
	assert.Contains(t, string(output), "f1: type=int64")
}

func TestColumnsCmd_Dataset(t *testing.T) {
	dir, err := ioutil.TempDir("", "arrowpipe-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	createTestDataset(t, dir)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"columns", dir})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	output, err := ioutil.ReadAll(buf)
	assert.NoError(t, err)
	assert.Equal(t, "1. f1: int64\n", string(output))
}

func TestStatsCmd_Dataset(t *testing.T) {
	dir, err := ioutil.TempDir("", "arrowpipe-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	createTestDataset(t, dir)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"stats", dir})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	output, err := ioutil.ReadAll(buf)
	assert.NoError(t, err)
	assert.Contains(t, string(output), "Total number of rows: 3")
}
