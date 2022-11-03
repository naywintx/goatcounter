// Copyright © Martin Tournoij – This file is part of GoatCounter and published
// under the terms of a slightly modified EUPL v1.2 license, which can be found
// in the LICENSE file or at https://license.goatcounter.com

package cron_test

import (
	"fmt"
	"testing"
	"time"

	"zgo.at/goatcounter/v2"
	"zgo.at/goatcounter/v2/gctest"
	"zgo.at/zstd/zjson"
	"zgo.at/zstd/ztest"
	"zgo.at/zstd/ztime"
)

func TestHitStats(t *testing.T) {
	ctx := gctest.DB(t)

	site := goatcounter.MustGetSite(ctx)
	now := time.Date(2019, 8, 31, 14, 42, 0, 0, time.UTC)

	// Store 3 pageviews for one session: two for "/asd" and one for "/zxc", all
	// on the same time.
	gctest.StoreHits(ctx, t, false, []goatcounter.Hit{
		{Site: site.ID, CreatedAt: now, Path: "/asd", Title: "aSd", FirstVisit: true},
		{Site: site.ID, CreatedAt: now, Path: "/asd/"}, // Trailing / should be sanitized and treated identical as /asd
		{Site: site.ID, CreatedAt: now, Path: "/zxc"},
	}...)

	check := func(wantT, want0, want1 string) {
		t.Helper()

		var stats goatcounter.HitLists
		displayUnique, more, err := stats.List(ctx,
			ztime.NewRange(now.Add(-1*time.Hour)).To(now.Add(1*time.Hour)),
			nil, nil, 10, false)
		if err != nil {
			t.Fatal(err)
		}

		gotT := fmt.Sprintf("%d %t", displayUnique, more)
		if wantT != gotT {
			t.Fatalf("wrong totals\nhave: %s\nwant: %s", gotT, wantT)
		}
		if len(stats) != 2 {
			t.Fatalf("len(stats) is not 2: %d", len(stats))
		}

		if d := ztest.Diff(string(zjson.MustMarshal(stats[0])), want0, ztest.DiffJSON); d != "" {
			t.Errorf("first wrong\n" + d)
		}

		if d := ztest.Diff(string(zjson.MustMarshal(stats[1])), want1, ztest.DiffJSON); d != "" {
			t.Errorf("second wrong\n" + d)
		}
	}

	check("1 false", `{
			"count_unique": 1,
			"path_id":      1,
			"path":         "/asd",
			"event":        false,
			"title":        "aSd",
			"max":          1,
			"stats": [{
				"day":           "2019-08-31",
				"hourly_unique": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0],
				"daily_unique":  1
			}]}
		`,
		`{
			"count_unique":  0,
			"path_id":       2,
			"path":          "/zxc",
			"event":         false,
			"title":         "",
			"max":           0,
			"stats": [{
				"day":           "2019-08-31",
				"hourly_unique": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
				"daily_unique":  0
			}]}`,
	)

	gctest.StoreHits(ctx, t, false, []goatcounter.Hit{
		{Site: site.ID, CreatedAt: now.Add(2 * time.Hour), Path: "/asd", Title: "aSd", FirstVisit: true},
		{Site: site.ID, CreatedAt: now.Add(2 * time.Hour), Path: "/asd", Title: "aSd"},
	}...)

	check("2 false", `{
			"count_unique":  2,
			"path_id":       1,
			"path":          "/asd",
			"event":         false,
			"title":         "aSd",
			"max":           1,
			"stats":[{
				"day":            "2019-08-31",
				"hourly_unique":  [0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0],
				"daily_unique":   2
		}]}`,
		`{
			"count_unique":  0,
			"path_id":       2,
			"path":          "/zxc",
			"event":         false,
			"title":         "",
			"max":           0,
			"stats":[{
				"day":            "2019-08-31",
				"hourly_unique":  [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
				"daily_unique":   0
		}]}`,
	)
}
