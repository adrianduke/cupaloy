package cupaloy

import (
	"encoding/json"
	"io"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func NewGocmpSnapshotter() *GocmpSnapshotter {
	return &GocmpSnapshotter{
		GocmpOptions: []cmp.Option{cmpopts.EquateApprox(0.00000000001, 0)},
	}
}

type GocmpSnapshotter struct {
	GocmpOptions []cmp.Option
}

func (s *GocmpSnapshotter) Snapshot(i ...interface{}) (Snap, error) {
	snap := []interface{}{}
	for _, obj := range i {
		snap = append(snap, obj)
	}

	return snap, nil
}
func (s *GocmpSnapshotter) WriteTo(writer io.Writer, target Snap) (int64, error) {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "	")

	return 0, encoder.Encode(target)
}
func (s *GocmpSnapshotter) ReadFrom(reader io.Reader) (Snap, error) {
	decoder := json.NewDecoder(reader)
	result := []interface{}{}
	err := decoder.Decode(&result)

	return result, err
}
func (s *GocmpSnapshotter) Diff(previous Snap, current Snap) string {
	return cmp.Diff(previous, current, s.GocmpOptions...)
}
