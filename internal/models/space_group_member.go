package models

const PublicAccessEmail = "@nyone"

type SpaceGroupMember struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type SpaceGroupMembers struct {
	Public  bool               `json:"public"`
	Members []SpaceGroupMember `json:"members"`
}
