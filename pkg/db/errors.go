package db

import "errors"

var (
	ErrRecordNotFound          = errors.New(`db: record not found`)
	ErrNoMoreRows              = errors.New(`db: no more rows in this result set`)
	ErrExpectingPointer        = errors.New(`db: argument must be an address`)
	ErrExpectingSlicePointer   = errors.New(`db: argument must be a slice address`)
	ErrExpectingSliceMapStruct = errors.New(`db: argument must be a slice address of maps or structs`)
	ErrExpectingMapOrStruct    = errors.New(`db: argument must be either a map or a struct`)
)
