package dots

import (
	"encoding/json"
	"errors"
)

type Power int

// create powers
const (
	DoAnything Power = iota
	CreateOwn
	WriteOwn
	ReadOwn
	DeleteOwn
)

var Powers = map[Power]string{
	DoAnything: "do anything for all",
	CreateOwn:  "Can create its own items",
	WriteOwn:   "Can edit its own items",
	ReadOwn:    "can list its own items",
	DeleteOwn:  "can delete its own items",
}

var ss = [...]string{"do_anything", "create_own", "write_own", "read_own", "delete_own"}

func (p Power) String() string {
	if int(p) > len(ss)-1 {
		return ""
	}
	return ss[p]
}

func (p Power) Bytes() []byte {
	if int(p) > len(ss)-1 {
		return nil
	}
	return []byte(ss[p])
}

func (p Power) Eq(other string) bool {
	return p.String() == other
}

func (p Power) Description() string {
	if desc, ok := Powers[p]; ok {
		return desc
	}
	return ""
}

func PowersContains(pp []Power, op Power) bool {
	for _, p := range pp {
		if p == op {
			return true
		}
	}
	return false
}

var PowerToManageOwn = []Power{CreateOwn, WriteOwn, ReadOwn}

func (p *Power) UnmarshalJSON(b []byte) error {
	if string(b) == "null" || string(b) == `""` {
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	for i, ps := range ss {
		if s == ps {
			*p = Power(i)
			return nil
		}
	}

	return errors.New("unmarshaling power")
}

func (p *Power) MarshalJSON() ([]byte, error) {
	bb := p.Bytes()
	if bb == nil {
		return nil, errors.New("marshaling power")
	}

	return bb, nil
}
