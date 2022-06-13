package model

type User struct {
	Id         string `json:"_id,omitempty"`
	Revision   string `json:"_rev,omitempty"`
	Username   string `json:"username"`
	FName      string `json:"firstname"`
	LName      string `json:"lastname"`
	IsAdmin    bool   `json:"isadmin"`
	DoorAccess Doors  `json:"dooraccess"`
}

type Doors map[string]bool

type UpdateAccessRequest struct {
	Username string `json:"username"`
	Doors    Doors  `json:"dooraccess"`
}

type DoorAuthenticate struct {
	Username   string `json:"username"`
	AccessDoor string `json:"accessdoor"`
}
