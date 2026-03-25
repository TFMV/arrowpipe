package engine

import (
    "context"

    "github.com/apache/arrow-go/v18/arrow"
)

// DataType is an interface for Arrow-like columnar data types.
type DataType interface {
    ID() arrow.Type
}

// Reader is the interface for reading columnar data.
type Reader interface {
    Read() (arrow.Record, error)
    Release()
}

// Writer is the interface for writing columnar data.
type Writer interface {
    Write(arrow.Record) error
    Close() error
}

// Stream is the interface for streaming columnar data.
type Stream interface {
    Reader
    Writer
}

// Hook is an interface for extending the engine's functionality.
type Hook interface {
    // Execute is called at a specific point in the engine's execution.
    Execute(context.Context, interface{}) error
}
