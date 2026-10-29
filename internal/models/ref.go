package models

type Ref[T ~string] struct {
	Id T `json:"id"`
}
