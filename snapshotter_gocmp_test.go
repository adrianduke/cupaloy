package cupaloy

import (
	"bytes"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func Test_Gocmp_JoinsMultipleObjectsIntoSlice(t *testing.T) {
	gocmp := NewGocmpSnapshotter()

	testObj1 := map[string]interface{}{
		"1": 1,
		"2": 2,
	}
	testObj2 := struct {
		A int
		B bool
		C []string
	}{
		A: 3,
		B: true,
		C: []string{"a", "b", "c"},
	}
	expectedObj := []interface{}{
		testObj1,
		testObj2,
	}

	returnedObj, err := gocmp.Snapshot(testObj1, testObj2)
	assert.NoError(t, err)

	assert.Equal(t, expectedObj, returnedObj)
}

func Test_Gocmp_WritesSnapshotToWriter(t *testing.T) {
	gocmp := NewGocmpSnapshotter()

	testObj1 := map[string]interface{}{
		"1": 1,
		"2": 2,
	}
	testObj2 := struct {
		A int
		B bool
		C []string
	}{
		A: 3,
		B: true,
		C: []string{"a", "b", "c"},
	}
	expectedStr := `[
	{
		"1": 1,
		"2": 2
	},
	{
		"A": 3,
		"B": true,
		"C": [
			"a",
			"b",
			"c"
		]
	}
]
`

	returnedObj, err := gocmp.Snapshot(testObj1, testObj2)
	assert.NoError(t, err)

	snapshot := &bytes.Buffer{}
	gocmp.WriteTo(snapshot, returnedObj)

	assert.Equal(t, expectedStr, snapshot.String())
}

func Test_Gocmp_ReadSnapshotToReader(t *testing.T) {
	gocmp := NewGocmpSnapshotter()

	reader := strings.NewReader(`[
	{
		"1": 1,
		"2": 2
	},
	{
		"A": 3,
		"B": true,
		"C": [
			"a",
			"b",
			"c"
		]
	}
]
`)

	expectedObj := []interface{}{
		map[string]interface{}{
			"1": 1.0,
			"2": 2.0,
		},
		map[string]interface{}{
			"A": 3.0,
			"B": true,
			"C": []interface{}{
				"a",
				"b",
				"c",
			},
		},
	}

	returnedObj, err := gocmp.ReadFrom(reader)
	assert.NoError(t, err)

	assert.Equal(t, expectedObj, returnedObj)
}

func Test_Gocmp_ReturnsDifferenceOf2Snapshots(t *testing.T) {
	gocmp := NewGocmpSnapshotter()

	tests := []struct {
		Name         string
		Snap1        Snap
		Snap2        Snap
		ExpectedDiff string
	}{
		{
			Name: "Slice ordering",
			Snap1: []interface{}{
				"1",
				"2",
				"3",
			},
			Snap2: []interface{}{
				"2",
				"1",
				"3",
			},
			ExpectedDiff: `[]interface{}{
-	string("1"),
+	string("2"),
-	string("2"),
+	string("1"),
	string("3"),
}
`,
		},
		{
			Name: "Float within margin",
			Snap1: []interface{}{
				0.9999999999999999,
				0.12345678901234,
				0.1234567890022,
			},
			Snap2: []interface{}{
				0.9999999999999999,
				0.12345678902234,
				0.123456789001,
			},
			ExpectedDiff: `[]interface{}{
	float64(0.9999999999999999),
-	float64(0.12345678901234),
+	float64(0.12345678902234),
	float64(0.123456789001),
}
`,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			diff := gocmp.Diff(test.Snap1, test.Snap2)

			// Normalise diff output, intermittently kept getting \u00a0 instead of whitespace \u0020
			transformer := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Zs)), norm.NFC)
			diff, _, err := transform.String(transformer, diff)
			assert.NoError(t, err)

			expectedDiff, _, err := transform.String(transformer, test.ExpectedDiff)
			assert.NoError(t, err)

			assert.Equal(t, expectedDiff, diff)
		})
	}

}
