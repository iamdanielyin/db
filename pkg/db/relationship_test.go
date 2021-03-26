package db

import "testing"

func TestHasOne(t *testing.T) {
	type B struct {
		ID  uint
		AID uint
	}
	// A has one B
	type A struct {
		ID uint
		B  B `rel:"has_one" localField:"ID" foreignField:"AID"`
	}
}

func TestHasMany(t *testing.T) {
	type B struct {
		ID  uint
		AID uint
	}
	// A has many B
	type A struct {
		ID uint
		Bs []B `rel:"has_many" localField:"ID" foreignField:"AID"`
	}
}

func TestRefOne(t *testing.T) {
	type B struct {
		ID uint
	}
	// A ref one B
	type A struct {
		ID  uint
		BID uint
		B   B `rel:"ref_one" localField:"ID" foreignField:"BID"`
	}
}

func TestRefManyNoSQL(t *testing.T) {
	type B struct {
		ID uint
	}
}

func TestRefMany(t *testing.T) {
	type B struct {
		ID uint
	}
	// A ref many B
	type C struct {
		ID  uint
		AID uint
		BID uint
	}
	type A struct {
		ID uint
		Bs []B `rel:"ref_many" localField:"ID=AID" foreignField:"ID=BID"`
	}
	// When the data source supports nested documents...
	type ANested struct {
		ID   uint
		BIDs []uint
		Bs   []B `rel:"ref_many" localField:"ID" foreignField:"BIDs"`
	}
}
