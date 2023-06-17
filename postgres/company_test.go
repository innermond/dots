package postgres_test

import (
	"context"
	"testing"

	"github.com/innermond/dots"
	"github.com/innermond/dots/postgres"
	"github.com/segmentio/ksuid"
)

func TestCompanyService_CreateCompany(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t, DSN)
		defer MustCloseDB(t, db)

		tid, err := getTid("user")
		if err != nil {
			t.Fatal(err)
		}

		tc := company{
			longname: "Long Name1",
			tin:      "tin1",
			rn:       "rn1",
		}

		s := postgres.NewCompanyService(db)
		u := dots.User{ID: tid, Powers: []dots.Power{dots.ReadOwn, dots.CreateOwn, dots.DeleteOwn}}
		ctx := dots.NewContextWithUser(context.Background(), &u)

		c, err, deleteDummy := createDummyCompany(ctx, t, tc, s)
		if err != nil {
			t.Fatalf("%v", err)
		}
		t.Logf("company created: %v\n", c)

		filterDel := dots.CompanyDelete{}
		filterDel.ID = &c.ID
		filterDel.Limit = 1
		// soft deletion
		n, err := s.DeleteCompany(ctx, filterDel)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		if n != 1 {
			t.Fatalf("expected to delete %d but deleted %d\n", 1, n)
		}
		t.Logf("company soft-deleted: %v\n", c.ID)

		filterFind := dots.CompanyFilter{ID: &c.ID}
		cc, n, err := s.FindCompany(ctx, filterFind)
		if err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatalf("expected 1 got %d\n", n)
		}
		t.Logf("found soft-deleted %v", cc[0].ID)

		// Hard deletion
		deleteDummy(cc[0].ID)
		t.Logf("company hard-deleted: %v\n", c.ID)

	})

	t.Run("Error characters that are empty or contains controls", func(t *testing.T) {
		db := MustOpenDB(t, DSN)
		defer MustCloseDB(t, db)

		tid, err := getTid("user")
		if err != nil {
			t.Fatal(err)
		}

		s := postgres.NewCompanyService(db)
		u := dots.User{ID: tid, Powers: []dots.Power{dots.ReadOwn, dots.CreateOwn, dots.DeleteOwn}}
		ctx := dots.NewContextWithUser(context.Background(), &u)

		tt := []company{
			{nil, "", "", ""},
			{nil, " ", " ", " "},
			{nil, "\t", "\n", "\b"},
			{nil, "longname", "very long\nname", "same\bhere"},
		}

		for i, tc := range tt {
			c, err, deleteDummy := createDummyCompany(ctx, t, tc, s)

			if err == nil {
				defer deleteDummy(c.ID)
				t.Fail()
				t.Logf("[%d] fail\n", i)
				continue
			}

			t.Logf("[%d] ok %v\n", i, err)
		}

	})

	t.Run("Error user sets TID", func(t *testing.T) {
		db := MustOpenDB(t, DSN)
		defer MustCloseDB(t, db)

		// tid of user
		tid, err := getTid("user")
		if err != nil {
			t.Fatal(err)
		}

		otid, err := getTid("other")
		if err != nil {
			t.Fatal(err)
		}

		tc := company{
			tid:      &otid,
			longname: "Long Name1",
			tin:      "tin1",
			rn:       "rn1",
		}

		s := postgres.NewCompanyService(db)
		u := dots.User{ID: tid, Powers: []dots.Power{dots.ReadOwn, dots.CreateOwn, dots.DeleteOwn}}
		ctx := dots.NewContextWithUser(context.Background(), &u)

		c, err, deleteDummy := createDummyCompany(ctx, t, tc, s)
		if err == nil {
			t.Logf("unexpected company creation: c.TID %v otid %v\n", c.TID, otid)
			deleteDummy(c.ID)
			t.Fatal("error unexpected")
		}
		t.Logf("error expected: %v\n", err)

		u = dots.User{ID: tid, Powers: []dots.Power{dots.DoAnything}}
		ctx = dots.NewContextWithUser(context.Background(), &u)
		c, err, deleteDummy = createDummyCompany(ctx, t, tc, s)
		if err != nil {
			t.Fatalf("unexpected error %v\n", err)
		}
		t.Logf("company created: c.TID %v\n", c.TID)
		deleteDummy(c.ID)

	})
}

type company struct {
	tid      *ksuid.KSUID
	longname string
	tin      string
	rn       string
}

func createDummyCompany(ctx context.Context, t *testing.T, tc company, s dots.CompanyService) (*dots.Company, error, func(int)) {
	c := &dots.Company{
		Longname: tc.longname,
		TIN:      tc.tin,
		RN:       tc.rn,
	}
	if tc.tid != nil {
		c.TID = *tc.tid
	}
	err := s.CreateCompany(ctx, c)

	deleteDummyCompany := func(cid int) {
		// do nothing if closure errored
		if err != nil {
			return
		}

		filterDel := dots.CompanyDelete{}
		filterDel.ID = &cid
		filterDel.Hard = true
		_, err := s.DeleteCompany(ctx, filterDel)
		if err != nil {
			t.Fail()
			t.Logf("%v\n", err)
		}
	}

	return c, err, deleteDummyCompany
}

func getTid(which string) (ksuid.KSUID, error) {
	if which == "user" {
		which = "2PH24UhBlN5tlYdAmpdwiyPuWgB"
	} else if which == "other" {
		which = "2PH25DxmohuFCf3w73fQSTLJeVO"
	} else {
		return ksuid.NewRandom()
	}

	tid := ksuid.KSUID{}
	// TODO for now it uses a real ID from db
	bb := []byte(which)
	err := tid.UnmarshalText(bb)
	return tid, err
}
