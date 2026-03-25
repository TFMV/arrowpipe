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

func TestSchemaCmd(t *testing.T) {
	// Create a dummy arrow file
	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "col1", Type: arrow.PrimitiveTypes.Int64},
			{Name: "col2", Type: arrow.BinaryTypes.String},
		},
		nil,
	)
	pool := memory.NewGoAllocator()

	tmpfile, err := ioutil.TempFile("", "arrowpipe-test-*.arrow")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	writer, err := storage.NewWriter(tmpfile, schema)
	assert.NoError(t, err)

	b := array.NewRecordBuilder(pool, schema)
	defer b.Release()

	b.Field(0).(*array.Int64Builder).Append(1)
	b.Field(1).(*array.StringBuilder).Append("a")

	rec := b.NewRecord()
	defer rec.Release()

	err = writer.Write(rec)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	// Run the schema command
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"schema", tmpfile.Name()})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Check the output
	expected := "schema:\n  fields: 2\n    - col1: type=int64\n    - col2: type=utf8\n"
	assert.Equal(t, expected, buf.String())
}
