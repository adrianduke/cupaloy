package cupaloy

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

func NewSpewSnapshotter() *SpewSnapshotter {
	return &SpewSnapshotter{
		ConfigState: spew.ConfigState{
			Indent:                  "  ",
			SortKeys:                true, // maps should be spewed in a deterministic order
			DisablePointerAddresses: true, // don't spew the addresses of pointers
			DisableCapacities:       true, // don't spew capacities of collections
			SpewKeys:                true, // if unable to sort map keys then spew keys to strings and sort those
		},
	}
}

type SpewSnapshotter struct {
	spew.ConfigState
}

func (s *SpewSnapshotter) Snapshot(i ...interface{}) (Snap, error) {
	snapshot := &bytes.Buffer{}
	for _, target := range i {
		switch typedTarget := target.(type) {
		case string:
			snapshot.WriteString(typedTarget)
			snapshot.WriteString("\n")
		case []byte:
			snapshot.Write(typedTarget)
			snapshot.WriteString("\n")
		default:
			s.Fdump(snapshot, typedTarget)
		}
	}

	return snapshot.String(), nil
}

func (s *SpewSnapshotter) WriteTo(writer io.Writer, target Snap) (int64, error) {
	return io.Copy(writer, strings.NewReader(target.(string)))
}
func (s *SpewSnapshotter) ReadFrom(reader io.Reader) (Snap, error) {
	snap, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return string(snap), nil
}
func (s *SpewSnapshotter) Diff(previous Snap, current Snap) string {
	if previous == nil {
		previous = ""
	}
	if current == nil {
		current = ""
	}

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(previous.(string)),
		B:        difflib.SplitLines(current.(string)),
		FromFile: "Previous",
		FromDate: "",
		ToFile:   "Current",
		ToDate:   "",
		Context:  1,
	})

	return diff
}
