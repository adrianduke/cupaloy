package cupaloy

import "io"

type Snap interface{}

type Snapshotter interface {
	Snapshot(...interface{}) (Snap, error)
	Diff(Snap, Snap) string

	ReadFrom(io.Reader) (Snap, error)
	WriteTo(io.Writer, Snap) (int64, error)
}
