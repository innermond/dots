package postgres_test

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/innermond/dots"
	"github.com/innermond/dots/postgres"
	"github.com/joho/godotenv"
	"github.com/segmentio/ksuid"
	"github.com/shopspring/decimal"
)

func TestDeed_Manage(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Fatal(err)
	}

	dsn := os.Getenv("DOTS_DSN")
	db := MustOpenDB(t, dsn)
	defer MustCloseDB(t, db)

	// create test user (it exists in db)
	uid, err := ksuid.Parse("2PH25DxmohuFCf3w73fQSTLJeVO")
	if err != nil {
		t.Fatalf("faking user: %v\n", err)
	}
	testuser := dots.User{
		ID:     uid,
		Powers: []dots.Power{dots.ReadOwn, dots.CreateOwn, dots.WriteOwn, dots.DeleteOwn},
	}
	// create context with user
	ctx := dots.NewContextWithUser(context.Background(), &testuser)

	entryService := postgres.NewEntryService(db)

	t.Log("create entry objects")

	f, err := os.Open("../testdata/entry.create.csv")
	if err != nil {
		t.Fatalf("Error opening CSV file: %v\n", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	i := -1

	entries := []dots.Entry{}

	for {

		line, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("Error reading CSV line: %v\n", err)
			return
		}

		i++

		// assume a header
		if i == 0 {
			continue
		}

		// parse line
		etid, err := strconv.Atoi(line[0])
		if err != nil {
			t.Fatalf("parsing entry type id: %v\n", err)
		}
		qty, err := strconv.ParseFloat(line[1], 64)
		if err != nil {
			t.Fatalf("parsing entry type quantity: %v\n", err)
		}
		cid, err := strconv.Atoi(line[2])
		if err != nil {
			t.Fatalf("parsing entry type company id: %v\n", err)
		}
		e := dots.Entry{0, etid, time.Time{}, qty, cid}

		t.Run(fmt.Sprintf("%d:", i), func(t *testing.T) {
			err := entryService.CreateEntry(ctx, &e)
			if err != nil {
				t.Fatalf("unexpected: %v\n", err)
			}
			entries = append(entries, e)
		})

	}

	t.Logf("created entries %v", entries)

	t.Log("delete all entries")

	for i, e := range entries {
		t.Run(fmt.Sprintf("%d:", i), func(t *testing.T) {
			_, err := entryService.DeleteEntry(ctx, e.ID, dots.EntryDelete{})
			if err != nil {
				t.Fatalf("unexpected: %v\n", err)
			}
		})
	}

	t.Log("deleted all entries")

	t.Log("resurect all entrie")

	for i, e := range entries {
		t.Run(fmt.Sprintf("%d:", i), func(t *testing.T) {
			_, err := entryService.DeleteEntry(ctx, e.ID, dots.EntryDelete{Resurect: true})
			if err != nil {
				t.Fatalf("unexpected: %v\n", err)
			}
		})
	}

	t.Log("resurected all entries")

	t.Log("updating entries")

	qty := 700.00
	cid := chooseRandomInt([]int{3, 5})
	upd := dots.EntryUpdate{Quantity: &qty, CompanyID: &cid}
	for i, e := range entries {
		t.Run(fmt.Sprintf("%d:", i), func(t *testing.T) {
			entry, err := entryService.UpdateEntry(ctx, e.ID, upd)
			if err != nil {
				t.Fatalf("unexpected: %v\n", err)
			}
			qty++
			entries[i] = *entry
			t.Logf("%v\n", entries)
		})
	}

	t.Log("updated entries")

	t.Log("create deed")

	deedService := postgres.NewDeedService(db)

	distribute := map[int]float64{}
	for _, e := range entries {
		distribute[e.ID] = e.Quantity * 0.01
	}
	deed := dots.Deed{0, cid, "Test title", 100, "buc", decimal.NewFromFloat(10.5), distribute, nil, nil}
	err = deedService.CreateDeed(ctx, &deed)
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}

	_, err = deedService.DeleteDeed(ctx, deed.ID, dots.DeedDelete{})
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}

	_, n, err := deedService.FindDeed(ctx, dots.DeedFilter{ID: &deed.ID})
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}
	if n != 0 {
		t.Fatalf("unexpected length %v\n", n)
	}

	_, err = deedService.DeleteDeed(ctx, deed.ID, dots.DeedDelete{Resurect: true})
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}

	_, n, err = deedService.FindDeed(ctx, dots.DeedFilter{ID: &deed.ID})
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}
	if n != 1 {
		t.Fatalf("unexpected length %v\n", n)
	}

	_, err = deedService.DeleteDeed(ctx, deed.ID, dots.DeedDelete{Undrain: true})
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}

	_, err = deedService.DeleteDeed(ctx, deed.ID, dots.DeedDelete{Resurect: true, Undrain: true})
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}

	_, err = deedService.DeleteDeed(ctx, deed.ID, dots.DeedDelete{Undrain: false})
	if err != nil {
		t.Fatalf("unexpected: %v\n", err)
	}

}

func chooseRandomInt(ii []int) int {
	rand.Seed(time.Now().UnixNano())
	l := len(ii)
	n := rand.Intn(l)
	return ii[n]
}
