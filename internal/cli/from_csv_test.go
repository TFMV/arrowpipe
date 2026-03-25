package cli

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/stretchr/testify/assert"
)

func TestFromCSVCmd(t *testing.T) {
	withTempFile(t, func(file *os.File) {
		// Prepare a CSV input
		csvInput := `a,b,c
1,hello,true
2,world,false`

		// Use the file for stdin
		_, err := file.WriteString(csvInput)
		assert.NoError(t, err)
		_, err = file.Seek(0, 0)
		assert.NoError(t, err)

		// Save original streams and restore them after the test
		oldIn := rootCmd.InOrStdin()
		oldOut := rootCmd.OutOrStdout()
		defer func() {
			rootCmd.SetIn(oldIn)
			rootCmd.SetOut(oldOut)
		}()

		rootCmd.SetIn(file)

		// Create a temporary file for the output
		outFile, err := ioutil.TempFile("", "arrow-test-out-")
		assert.NoError(t, err)
		defer os.Remove(outFile.Name())
		defer outFile.Close()

		rootCmd.SetOut(outFile)

		// Run the command
		rootCmd.SetArgs([]string{"from-csv"})
		err = rootCmd.Execute()
		assert.NoError(t, err)

		// Read the arrow output from the output file
		_, err = outFile.Seek(0, 0)
		assert.NoError(t, err)

		rdr, err := ipc.NewFileReader(outFile)
		assert.NoError(t, err)
		defer rdr.Close()

		// Check the schema
		schema := rdr.Schema()
		assert.Equal(t, 3, schema.NumFields())
		assert.Equal(t, "a", schema.Field(0).Name)
		assert.Equal(t, "b", schema.Field(1).Name)
		assert.Equal(t, "c", schema.Field(2).Name)

		// Check the data
		assert.Equal(t, 1, rdr.NumRecords())

		rec, err := rdr.Record(0)
		assert.NoError(t, err)
		defer rec.Release()

		assert.Equal(t, int64(2), rec.NumRows())
		assert.Equal(t, "[1 2]", rec.Column(0).String())
		assert.Equal(t, `["hello" "world"]`, rec.Column(1).String())
		assert.Equal(t, "[true false]", rec.Column(2).String())
	})
}
