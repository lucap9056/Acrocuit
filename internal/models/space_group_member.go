package models

type SpaceGroupMember struct {
	Email string `json:"email"`
	Role  string `json:"role" tstype:"SpaceGroupRole"`
}

type SpaceGroupMembers struct {
	Public  bool               `json:"public"`
	Members []SpaceGroupMember `json:"members"`
}
