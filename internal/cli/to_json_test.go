package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
)

func TestToJSONCmd(t *testing.T) {
	// Create a sample Arrow record
	mem := memory.NewGoAllocator()
	fields := []arrow.Field{
		{Name: "a", Type: arrow.PrimitiveTypes.Int64},
		{Name: "b", Type: arrow.BinaryTypes.String},
	}
	schema := arrow.NewSchema(fields, nil)

	ib := array.NewInt64Builder(mem)
	ib.Append(1)
	ib.Append(2)
	iarr := ib.NewInt64Array()
	defer iarr.Release()
	defer ib.Release()

	sb := array.NewStringBuilder(mem)
	sb.Append("hello")
	sb.Append("world")
	sarr := sb.NewStringArray()
	defer sarr.Release()
	defer sb.Release()

	rec := array.NewRecord(schema, []arrow.Array{iarr, sarr}, 2)
	defer rec.Release()

	// Write the record to a buffer in IPC format
	var buf bytes.Buffer
	w, err := storage.NewWriter(&buf, schema)
	assert.NoError(t, err)
	assert.NoError(t, w.Write(rec))
	assert.NoError(t, w.Close())

	// Save original streams and restore them after the test
	oldIn := rootCmd.InOrStdin()
	oldOut := rootCmd.OutOrStdout()
	defer func() {
		rootCmd.SetIn(oldIn)
		rootCmd.SetOut(oldOut)
	}()

	// Use the buffer as stdin
	rootCmd.SetIn(&buf)

	// Create a temporary file for the output
	outFile, err := ioutil.TempFile("", "arrow-test-out-")
	assert.NoError(t, err)
	defer os.Remove(outFile.Name())
	defer outFile.Close()

	rootCmd.SetOut(outFile)

	// Run the command
	rootCmd.SetArgs([]string{"to-json"})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Read the output and check it
	output, err := ioutil.ReadFile(outFile.Name())
	assert.NoError(t, err)

	expectedJSON := `{"a":"1","b":"hello"}
{"a":"2","b":"world"}
`
	assert.Equal(t, expectedJSON, string(output))
}
