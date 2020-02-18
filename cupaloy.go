package cupaloy

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/adrianduke/cupaloy/internal"
)

type Cupaloy struct {
	*Config
	Snapshotter
}

// New constructs a new, configured instance of cupaloy using the given
// Configurators applied to the default config.
func New(configurators ...Configurator) *Cupaloy {
	return &Cupaloy{
		Config:      NewDefaultConfig().WithOptions(configurators...),
		Snapshotter: NewSpewSnapshotter(),
	}
}

// Snapshot calls Snapshotter.Snapshot with the global config.
func Snapshot(i ...interface{}) error {
	return Global.snapshot(getNameOfCaller(), i...)
}

// SnapshotMulti calls Snapshotter.SnapshotMulti with the global config.
func SnapshotMulti(snapshotID string, i ...interface{}) error {
	snapshotName := fmt.Sprintf("%s-%s", getNameOfCaller(), snapshotID)
	return Global.snapshot(snapshotName, i...)
}

// SnapshotT calls Snapshotter.SnapshotT with the global config.
func SnapshotT(t TestingT, i ...interface{}) {
	t.Helper()
	Global.SnapshotT(t, i...)
}

// Snapshot compares the given variable to its previous value stored on the filesystem.
// An error containing a diff is returned if the snapshots do not match, or if a new
// snapshot was created.
//
// Snapshot determines the snapshot file automatically from the name of the calling function.
// As a result it can be called at most once per function. If you want to call Snapshot
// multiple times in a function, if possible, instead collect the values and call Snapshot
// with all values at once. Otherwise see SnapshotMulti.
//
// If using snapshots in tests, prefer the SnapshotT function which fails the test
// directly, rather than requiring your to remember to check the error.
func (c *Cupaloy) Snapshot(i ...interface{}) error {
	return c.snapshot(getNameOfCaller(), i...)
}

// SnapshotMulti is similar to Snapshot but can be called multiple times from the
// same function. This is possible by providing a unique id for each snapshot which is
// appended to the function name to form the snapshot name.
func (c *Cupaloy) SnapshotMulti(snapshotID string, i ...interface{}) error {
	snapshotName := fmt.Sprintf("%s-%s", getNameOfCaller(), snapshotID)
	return c.snapshot(snapshotName, i...)
}

// SnapshotT compares the given variable to the its previous value stored on the filesystem.
// The current test is failed (with error containing a diff) if the values do not match, or
// if a new snapshot was created.
//
// SnapshotT determines the snapshot file automatically from the name of the test (using
// the t.Name() function). As a result, SnapshotT can be called at most once per test.
// If you want to call SnapshotT multiple times in a test, if possible, instead collect the
// values and call SnapshotT with all values at once. Alternatively, use sub-tests and call
// SnapshotT once in each.
//
// If using snapshots in tests, SnapshotT is preferred over Snapshot and SnapshotMulti.
func (c *Cupaloy) SnapshotT(t TestingT, i ...interface{}) {
	t.Helper()
	if t.Failed() {
		return
	}

	snapshotName := strings.Replace(t.Name(), "/", "-", -1)
	err := c.snapshot(snapshotName, i...)
	if err != nil {
		if c.fatalOnMismatch {
			t.Fatal(err)
			return
		}
		t.Error(err)
	}
}

func (c *Cupaloy) WithOptions(configurators ...Configurator) *Cupaloy {
	return &Cupaloy{
		Config:      c.Config.WithOptions(configurators...),
		Snapshotter: NewSpewSnapshotter(),
	}
}

func (c *Cupaloy) snapshot(snapshotName string, i ...interface{}) error {
	snapshot, err := c.Snapshotter.Snapshot(i...)
	if err != nil {
		return err
	}

	buf, err := os.Open(c.snapshotFilePath(snapshotName))
	if os.IsNotExist(err) {
		if c.createNewAutomatically {
			return c.updateSnapshot(snapshotName, nil, snapshot)
		}

		return internal.ErrNoSnapshot{Name: snapshotName}
	} else if err != nil {
		return err
	}

	prevSnapshot, err := c.Snapshotter.ReadFrom(buf)
	if err != nil {
		return err
	}

	diff := c.Snapshotter.Diff(prevSnapshot, snapshot)
	if diff == "" || takeV1Snapshot(i...) == prevSnapshot {
		// previous snapshot matches current value
		return nil
	}

	if c.shouldUpdate() {
		// updates snapshot to current value and upgrades snapshot format
		return c.updateSnapshot(snapshotName, prevSnapshot, snapshot)
	}

	return internal.ErrSnapshotMismatch{
		Diff: diff,
	}
}

func (c *Cupaloy) updateSnapshot(snapshotName string, prevSnapshot, snapshot Snap) error {
	// check that subdirectory exists before writing snapshot
	err := os.MkdirAll(c.subDirName, os.ModePerm)
	if err != nil {
		return errors.New("could not create snapshots directory")
	}

	snapshotFile := c.snapshotFilePath(snapshotName)
	_, err = os.Stat(snapshotFile)
	isNewSnapshot := os.IsNotExist(err)

	f, err := os.Create(snapshotFile)
	if err != nil {
		return err
	}
	defer f.Close()

	c.Snapshotter.WriteTo(f, snapshot)

	if !c.failOnUpdate {
		//TODO: should a warning still be printed here?
		return nil
	}

	snapshotDiff := c.Snapshotter.Diff(prevSnapshot, snapshot)

	if isNewSnapshot {
		return internal.ErrSnapshotCreated{
			Name:     snapshotName,
			Contents: snapshot,
		}
	}

	return internal.ErrSnapshotUpdated{
		Name: snapshotName,
		Diff: snapshotDiff,
	}
}
