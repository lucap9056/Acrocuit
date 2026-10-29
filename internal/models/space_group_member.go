package models

const (
	PublicAccessEmail = "@nyone"
	SingleUserEmail   = "owner@localhost"
)

type SpaceGroupMember struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type SpaceGroupMembers struct {
	Public  bool               `json:"public"`
	Members []SpaceGroupMember `json:"members"`
}
