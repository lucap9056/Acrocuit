package models

const BASE_POSITION = "position"

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}
