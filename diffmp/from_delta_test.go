package diffmp

import (
	"testing"

	"github.com/stretchrcom/testify/assert"
)

func TestFromDelta(t *testing.T) {
	// Convert a diff into delta string.
	diffs := []Diff{
		diffNoop("jump"),
		diffDel("s"),
		diffIns("ed"),
		diffNoop(" over "),
		diffDel("the"),
		diffIns("a"),
		diffNoop(" lazy"),
		diffIns("old dog"),
	}

	text1 := DiffText1(diffs)
	assert.Equal(t, "jumps over the lazy", text1)

	delta := DiffToDelta(diffs)
	assert.Equal(t, "=4\t-1\t+ed\t=6\t-3\t+a\t=5\t+old dog", delta)

	// Convert delta string into a diff.
	_seq1, err := FromDelta(text1, delta)
	assertDiffEqual(t, diffs, _seq1)

	// Generates error (19 < 20).
	_, err = FromDelta(text1+"x", delta)
	if err == nil {
		t.Fatal("too long")
	}

	// Generates error (19 > 18).
	_, err = FromDelta(text1[1:], delta)
	if err == nil {
		t.Fatal("too short")
	}

	// Generates error (%xy invalid URL escape).
	_, err = FromDelta("", "+%c3%xy")
	if err == nil {
		assert.Fail(t, "expected Invalid URL escape")
	}

	// Generates error (invalid utf8).
	_, err = FromDelta("", "+%c3xy")
	if err == nil {
		assert.Fail(t, "expected Invalid utf8")
	}

	// Test deltas with special characters.
	diffs = []Diff{
		diffNoop("\u0680 \x00 \t %"),
		diffDel("\u0681 \x01 \n ^"),
		diffIns("\u0682 \x02 \\ |"),
	}
	text1 = DiffText1(diffs)
	assert.Equal(t, "\u0680 \x00 \t %\u0681 \x01 \n ^", text1)

	delta = DiffToDelta(diffs)
	// Lowercase, due to UrlEncode uses lower.
	assert.Equal(t, "=7\t-7\t+%DA%82 %02 %5C %7C", delta)

	_res1, err := FromDelta(text1, delta)
	if err != nil {
		t.Fatal(err)
	}
	assertDiffEqual(t, diffs, _res1)

	// Verify pool of unchanged characters.
	diffs = []Diff{
		{Insert, "A-Z a-z 0-9 - _ . ! ~ * ' ( ) ; / ? : @ & = + $ , # "}}
	text2 := DiffText2(diffs)
	assert.Equal(
		t, "A-Z a-z 0-9 - _ . ! ~ * ' ( ) ; / ? : @ & = + $ , # ",
		text2, "unchanged characters",
	)

	delta = DiffToDelta(diffs)
	assert.Equal(
		t, "+A-Z a-z 0-9 - _ . ! ~ * ' ( ) ; / ? : @ & = + $ , # ",
		delta, "unchanged characters",
	)

	// Convert delta string into a diff.
	_res2, _ := FromDelta("", delta)
	assertDiffEqual(t, diffs, _res2)
}
