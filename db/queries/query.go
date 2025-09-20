package queries

import "fmt"

var (
	ErrNoRowFound         = fmt.Errorf("no row found")
	ErrNoRowInserted      = fmt.Errorf("no row inserted")
	ErrNoRowFoundToUpdate = fmt.Errorf("no row found to update")
)
