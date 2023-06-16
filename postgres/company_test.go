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

    var err error

    tid := ksuid.KSUID{}
    // TODO for now it uses a real ID from db
    bb := []byte("2PH24UhBlN5tlYdAmpdwiyPuWgB")
    err = tid.UnmarshalText(bb)
    if err != nil {
      t.Fatal(err)
    }
    c := &dots.Company{
      TID: tid,
      Longname: "Long Name1",
      TIN: "tin1",
      RN: "rn1",
    }

    s := postgres.NewCompanyService(db)
    u := dots.User{ID: tid, Powers: []dots.Power{dots.ReadOwn, dots.CreateOwn, dots.DeleteOwn}}
    ctx := dots.NewContextWithUser(context.Background(), &u)
    err = s.CreateCompany(ctx, c)
    if err != nil {
      t.Fatalf("%v", err)
    }
    t.Logf("company created: %v\n", c)

    filterDel := dots.CompanyDelete{}
    filterDel.CompanyFilter.ID = &c.ID 
    filterDel.CompanyFilter.Limit = 1
    // soft deletion
    n, err := s.DeleteCompany(ctx, filterDel)
    if err != nil {
      t.Fatalf("%v\n", err)
    }
    if n != 1 {
      t.Fatalf("expected to delete %d but deleted %d\n", 1, n)
    }

    filterFind := dots.CompanyFilter {ID: &c.ID}
    cc, n, err := s.FindCompany(ctx, filterFind)
    if err != nil {
      t.Fatal(err)
    }
    if n != 1 {
      t.Fatalf("expected 1 got %d\n", n)
    }
    t.Logf("deleted %v", cc[0].ID)

    // Hard deletion
    // reset
    filterDel = dots.CompanyDelete{}
    // init
    filterDel.ID = &cc[0].ID
    filterDel.Hard = true
    _, err = s.DeleteCompany(ctx, filterDel)
    if err != nil {
      t.Fatalf("%v\n", err)
    }

  })
}
