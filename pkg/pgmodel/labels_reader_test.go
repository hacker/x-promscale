package pgmodel

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestLabelsReaderLabelsNames(t *testing.T) {
	testCases := []struct {
		name        string
		expectedRes []string
		sqlQueries  []sqlQuery
	}{
		{
			name: "Error on query",
			sqlQueries: []sqlQuery{
				{
					sql:     "SELECT distinct key from _prom_catalog.label",
					args:    []interface{}(nil),
					results: rowResults{},
					err:     fmt.Errorf("some error"),
				},
			},
		}, {
			name: "Error on scanning values",
			sqlQueries: []sqlQuery{
				{
					sql:     "SELECT distinct key from _prom_catalog.label",
					args:    []interface{}(nil),
					results: rowResults{{1}},
				},
			},
		}, {
			name: "Empty result, is ok",
			sqlQueries: []sqlQuery{
				{
					sql:     "SELECT distinct key from _prom_catalog.label",
					args:    []interface{}(nil),
					results: rowResults{},
				},
			},
			expectedRes: []string{},
		}, {
			name: "Result should be sorted",
			sqlQueries: []sqlQuery{
				{
					sql:     "SELECT distinct key from _prom_catalog.label",
					args:    []interface{}(nil),
					results: rowResults{{"b"}, {"a"}},
				},
			},
			expectedRes: []string{"a", "b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := newSqlRecorder(tc.sqlQueries, t)
			reader := labelsReader{conn: mock}
			res, err := reader.LabelNames()

			var expectedErr error
			for _, q := range tc.sqlQueries {
				if q.err != nil {
					expectedErr = err
					break
				}
			}

			if tc.name == "Error on scanning values" {
				if err.Error() != "wrong value type int" {
					expectedErr = fmt.Errorf("wrong value type int")
					t.Errorf("unexpected error\n got: %v\n expected: %v", err, expectedErr)
					return
				}
			} else if expectedErr != err {
				t.Errorf("unexpected error\n got: %v\n expected: %v", err, expectedErr)
				return
			}

			outputIsSorted := sort.SliceIsSorted(res, func(i, j int) bool {
				return res[i] < res[j]
			})
			if !outputIsSorted {
				t.Errorf("returned label names %v are not sorted", res)
			}

			if !reflect.DeepEqual(tc.expectedRes, res) {
				t.Errorf("expected: %v, got: %v", tc.expectedRes, res)
			}
		})
	}
}

func TestLabelsReaderLabelsValues(t *testing.T) {
	testCases := []struct {
		name        string
		expectedRes []string
		sqlQueries  []sqlQuery
	}{
		{
			name: "Error on query",
			sqlQueries: []sqlQuery{
				{
					sql:     "SELECT value from _prom_catalog.label WHERE key = $1",
					args:    []interface{}{"m"},
					results: rowResults{},
					err:     fmt.Errorf("some error"),
				},
			},
		}, {
			name: "Error on scanning values",
			sqlQueries: []sqlQuery{
				{
					sql:     "SELECT value from _prom_catalog.label WHERE key = $1",
					args:    []interface{}{"m"},
					results: rowResults{{1}},
				},
			},
		}, {
			name: "Empty result, is ok",
			sqlQueries: []sqlQuery{
				{
					sql:     "SELECT value from _prom_catalog.label WHERE key = $1",
					args:    []interface{}{"m"},
					results: rowResults{},
				},
			},
			expectedRes: []string{},
		}, {
			name: "Result should be sorted",
			sqlQueries: []sqlQuery{
				{
					sql:     "SELECT value from _prom_catalog.label WHERE key = $1",
					args:    []interface{}{"m"},
					results: rowResults{{"b"}, {"a"}},
				},
			},
			expectedRes: []string{"a", "b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := newSqlRecorder(tc.sqlQueries, t)
			querier := labelsReader{conn: mock}
			res, err := querier.LabelValues("m")

			var expectedErr error
			for _, q := range tc.sqlQueries {
				if q.err != nil {
					expectedErr = err
					break
				}
			}

			if tc.name == "Error on scanning values" {
				if err.Error() != "wrong value type int" {
					expectedErr = fmt.Errorf("wrong value type int")
					t.Errorf("unexpected error\n got: %v\n expected: %v", err, expectedErr)
					return
				}
			} else if expectedErr != err {
				t.Errorf("unexpected error\n got: %v\n expected: %v", err, expectedErr)
				return
			}

			outputIsSorted := sort.SliceIsSorted(res, func(i, j int) bool {
				return res[i] < res[j]
			})
			if !outputIsSorted {
				t.Errorf("returned label names %v are not sorted", res)
			}

			if !reflect.DeepEqual(tc.expectedRes, res) {
				t.Errorf("expected: %v, got: %v", tc.expectedRes, res)
			}
		})
	}
}
