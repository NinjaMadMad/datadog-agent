// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package metrics

import (
	// stdlib
	"fmt"
	"testing"

	// 3p
	"github.com/stretchr/testify/assert"
)

// AssertPointsEqual evaluate if two list of point are equal (order doesn't matters).
func AssertPointsEqual(t *testing.T, expected, actual []Point) {
	if assert.Equal(t, len(expected), len(actual)) {
		for _, point := range expected {
			assert.Contains(t, actual, point)
		}
	}
}

// AssertTagsEqual evaluate if two list of tags are equal (the order doesn't matters).
func AssertTagsEqual(t *testing.T, expected, actual []string) {
	if assert.Equal(t, len(expected), len(actual), fmt.Sprintf("Unexpected number of tags: expected %s, actual: %s", expected, actual)) {
		for _, tag := range expected {
			assert.Contains(t, actual, tag)
		}
	}
}

// AssertSerieEqual evaluate if two are equal.
func AssertSerieEqual(t *testing.T, expected, actual *Serie) {
	assert.Equal(t, expected.Name, actual.Name)
	if expected.Tags != nil {
		assert.NotNil(t, actual.Tags)
		AssertTagsEqual(t, expected.Tags, actual.Tags)
	}
	assert.Equal(t, expected.Host, actual.Host)
	assert.Equal(t, expected.MType, actual.MType)
	assert.Equal(t, expected.Interval, actual.Interval)
	assert.Equal(t, expected.SourceTypeName, actual.SourceTypeName)
	if !expected.ContextKey.IsZero() {
		// Only test the contextKey if it's set in the expected Serie
		assert.Equal(t, expected.ContextKey, actual.ContextKey)
	}
	assert.Equal(t, expected.NameSuffix, actual.NameSuffix)
	AssertPointsEqual(t, expected.Points, actual.Points)
}

// // AssertSketchSeriesEqual checks whether two SketchSeries are equal
// func AssertSketchSeriesEqual(t *testing.T, expected, actual *SketchSeries) {
// 	assert.Equal(t, expected.Name, actual.Name)
// 	if expected.Tags != nil {
// 		assert.NotNil(t, actual.Tags)
// 		AssertTagsEqual(t, expected.Tags, actual.Tags)
// 	}
// 	assert.Equal(t, expected.Host, actual.Host)
// 	assert.Equal(t, expected.Interval, actual.Interval)
// 	if !expected.ContextKey.IsZero() {
// 		assert.Equal(t, expected.ContextKey, actual.ContextKey)
// 	}
// 	if expected.Sketches != nil {
// 		assert.NotNil(t, actual.Sketches)
// 		AssertSketchesEqual(t, expected.Sketches, actual.Sketches)
// 	}
// }

// // AssertSketchesEqual checks whether two Sketch slices are equal
// func AssertSketchesEqual(t *testing.T, expected, actual []Sketch) {
// 	if assert.Equal(t, len(expected), len(actual)) {
// 		sort.Slice(actual, func(i, j int) bool {
// 			return actual[i].Ts < actual[j].Ts
// 		})
// 		for i, sketch := range expected {
// 			assert.Equal(t, sketch, actual[i])
// 		}
// 	}
// }
