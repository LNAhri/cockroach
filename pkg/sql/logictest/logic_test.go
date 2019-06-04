// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License included
// in the file licenses/BSL.txt and at www.mariadb.com/bsl11.
//
// Change Date: 2022-10-01
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt and at
// https://www.apache.org/licenses/LICENSE-2.0

package logictest

import (
	"go/build"
	"os"
	"path/filepath"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
)

// TestLogic runs logic tests that were written by hand to test various
// CockroachDB features. The tests use a similar methodology to the SQLLite
// Sqllogictests. All of these tests should only verify correctness of output,
// and not how that output was derived. Therefore, these tests can be run
// using the heuristic planner, the cost-based optimizer, or even run against
// Postgres to verify it returns the same logical results.
//
// See the comments in logic.go for more details.
func TestLogic(t *testing.T) {
	defer leaktest.AfterTest(t)()
	RunLogicTest(t, "testdata/logic_test/[^.]*")
}

// TestPlannerLogic tests the heuristic planner by running EXPLAIN and SHOW
// TRACE queries that show the plan that was produced. These tests are split
// off from the TestLogic tests because the expected output is specific to how
// the planner works. The cost-based optimizer will often return different
// results for the same EXPLAIN statement, as it often chooses different ways
// to execute the same logical query. Note that the cost-based optimizer tests
// are housed in the various sql/opt packages.
func TestPlannerLogic(t *testing.T) {
	defer leaktest.AfterTest(t)()
	RunLogicTest(t, "testdata/planner_test/[^.]*")
}

// TestSqlLiteLogic runs the subset of SqlLite logic tests that do not require
// support for correlated subqueries. The heuristic planner does not support
// correlated subqueries, so until that is fully deprecated, it can only run
// this subset.
//
// See the comments for runSQLLiteLogicTest for more detail on these tests.
func TestSqlLiteLogic(t *testing.T) {
	defer leaktest.AfterTest(t)()
	runSQLLiteLogicTest(t,
		"/test/index/between/*/*.test",
		"/test/index/commute/*/*.test",
		"/test/index/delete/*/*.test",
		"/test/index/in/*/*.test",
		"/test/index/orderby/*/*.test",
		"/test/index/orderby_nosort/*/*.test",
		"/test/index/view/*/*.test",

		// TODO(pmattis): Incompatibilities in numeric types.
		// For instance, we type SUM(int) as a decimal since all of our ints are
		// int64.
		// "/test/random/expr/*.test",

		// TODO(pmattis): We don't support unary + on strings.
		// "/test/index/random/*/*.test",
		// "/test/random/aggregates/*.test",
		// "/test/random/groupby/*.test",
		// "/test/random/select/*.test",
	)
}

// TestSqlLiteCorrelatedLogic runs the subset of SqlLite logic tests that
// require support for correlated subqueries. The cost-based optimizer has this
// support, whereas the heuristic planner does not.
//
// See the comments for runSQLLiteLogicTest for more detail on these tests.
func TestSqlLiteCorrelatedLogic(t *testing.T) {
	defer leaktest.AfterTest(t)()
	runSQLLiteLogicTest(t,
		"/test/select1.test",
		"/test/select2.test",
		"/test/select3.test",
		"/test/select4.test",

		// TODO(andyk): No support for join ordering yet, so this takes too long.
		// "/test/select5.test",
	)
}

// runSQLLiteLogicTest runs logic tests from CockroachDB's fork of sqllogictest:
//
//   https://www.sqlite.org/sqllogictest/doc/trunk/about.wiki
//
// This fork contains many generated tests created by the SqlLite project that
// ensure the tested SQL database returns correct statement and query output.
// The logic tests are reasonably independent of the specific dialect of each
// database so that they can be retargeted. In fact, the expected output for
// each test can be generated by one database and then used to verify the output
// of another database.
//
// By default, these tests are skipped, unless the `bigtest` flag is specified.
// The reason for this is that these tests are contained in another repo that
// must be present on the machine, and because they take a long time to run.
//
// See the comments in logic.go for more details.
func runSQLLiteLogicTest(t *testing.T, globs ...string) {
	if !*bigtest {
		t.Skip("-bigtest flag must be specified to run this test")
	}

	logicTestPath := build.Default.GOPATH + "/src/github.com/cockroachdb/sqllogictest"
	if _, err := os.Stat(logicTestPath); os.IsNotExist(err) {
		fullPath, err := filepath.Abs(logicTestPath)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("unable to find sqllogictest repo: %s\n"+
			"git clone https://github.com/cockroachdb/sqllogictest %s",
			logicTestPath, fullPath)
		return
	}

	// Prefix the globs with the logicTestPath.
	prefixedGlobs := make([]string, len(globs))
	for i, glob := range globs {
		prefixedGlobs[i] = logicTestPath + glob
	}

	RunLogicTest(t, prefixedGlobs...)
}
