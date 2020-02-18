package cupaloy

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

//go:generate $GOPATH/bin/mockery -output=examples -outpkg=examples_test -testonly -name=TestingT

// TestingT is a subset of the interface testing.TB allowing it to be mocked in tests.
type TestingT interface {
	Helper()
	Failed() bool
	Error(args ...interface{})
	Fatal(args ...interface{})
	Name() string
}

func getNameOfCaller() string {
	pc, _, _, _ := runtime.Caller(2) // first caller is the caller of this function, we want the caller of our caller
	fullPath := runtime.FuncForPC(pc).Name()
	packageFunctionName := filepath.Base(fullPath)

	return strings.Replace(packageFunctionName, ".", "-", -1)
}

func envVariableSet(envVariable string) bool {
	_, varSet := os.LookupEnv(envVariable)
	return varSet
}

func (c *Config) snapshotFilePath(testName string) string {
	return filepath.Join(c.subDirName, testName+c.snapshotFileExtension)
}

// Legacy snapshot format where all items were spewed
func takeV1Snapshot(i ...interface{}) string {
	spewConfig := spew.ConfigState{
		Indent:                  "  ",
		SortKeys:                true, // maps should be spewed in a deterministic order
		DisablePointerAddresses: true, // don't spew the addresses of pointers
		DisableCapacities:       true, // don't spew capacities of collections
		SpewKeys:                true, // if unable to sort map keys then spew keys to strings and sort those
	}

	return spewConfig.Sdump(i...)
}
