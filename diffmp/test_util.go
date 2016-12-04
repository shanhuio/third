package diffmp

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func caller() string {
	if _, _, line, ok := runtime.Caller(2); ok {
		return fmt.Sprintf("(actual-line %v) ", line)
	}
	return ""
}

func pretty(diffs []Diff) string {
	w := new(bytes.Buffer)
	for i, d := range diffs {
		fmt.Fprintf(w, "%d %s %s\n", i, opStr(d.Type), d.Text)
	}
	return w.String()
}

func assertDiffEqual(t *testing.T, seq1, seq2 []Diff) {
	if a, b := len(seq1), len(seq2); a != b {
		t.Errorf(
			"%v\nseq1:\n%v\nseq2:\n%v",
			caller(), pretty(seq1), pretty(seq2),
		)
		t.Errorf(
			"%v Sequences of different length: %v != %v",
			caller(), a, b,
		)
		return
	}

	for i := range seq1 {
		if a, b := seq1[i], seq2[i]; a != b {
			t.Errorf(
				"%v\nseq1:\n%v\nseq2:\n%v",
				caller(), pretty(seq1), pretty(seq2),
			)
			t.Errorf("%v %v != %v", caller(), a, b)
			return
		}
	}
}

func assertMapEqual(t *testing.T, seq1, seq2 interface{}) {
	v1 := reflect.ValueOf(seq1)
	k1 := v1.Kind()
	v2 := reflect.ValueOf(seq2)
	k2 := v2.Kind()

	if k1 != reflect.Map || k2 != reflect.Map {
		t.Fatalf("%v Parameters are not maps", caller())
	} else if v1.Len() != v2.Len() {
		t.Fatalf(
			"%v Maps of different length: %v != %v",
			caller(), v1.Len(), v2.Len(),
		)
	}

	keys1, keys2 := v1.MapKeys(), v2.MapKeys()

	if len(keys1) != len(keys2) {
		t.Fatalf("%v Maps of different length", caller())
	}

	for _, key1 := range keys1 {
		a := v1.MapIndex(key1).Interface()
		b := v2.MapIndex(key1).Interface()
		if a != b {
			t.Fatalf(
				"%v Different key/value in Map: %v != %v",
				caller(), a, b,
			)
		}
	}

	for _, key2 := range keys2 {
		a := v1.MapIndex(key2).Interface()
		b := v2.MapIndex(key2).Interface()
		if a != b {
			t.Fatalf(
				"%v Different key/value in Map: %v != %v",
				caller(), a, b,
			)
		}
	}
}

func assertStrEqual(t *testing.T, seq1, seq2 []string) {
	if a, b := len(seq1), len(seq2); a != b {
		t.Fatalf(
			"%v Sequences of different length: %v != %v",
			caller(), a, b,
		)
	}

	for i := range seq1 {
		if a, b := seq1[i], seq2[i]; a != b {
			t.Fatalf("%v %v != %v", caller(), a, b)
		}
	}
}

func diffRebuildtexts(diffs []Diff) []string {
	text := []string{"", ""}
	for _, myDiff := range diffs {
		if myDiff.Type != Insert {
			text[0] += myDiff.Text
		}
		if myDiff.Type != Delete {
			text[1] += myDiff.Text
		}
	}
	return text
}
