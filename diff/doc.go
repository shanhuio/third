// Package diff is a partial port of Python difflib module.
//
// It provides tools to compare sequences of strings and generate textual diffs.
//
// The following class and functions have been ported:
//
// - SequenceMatcher
// - unified_diff
// - context_diff
//
// Getting unified diffs was the main goal of the port. Keep in mind this code
// is mostly suitable to output text differences in a human friendly way, there
// are no guarantees generated diffs are consumable by patch(1).
package diff
