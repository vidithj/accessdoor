package api

import (
	"accessdoor/model"
	eventmodel "events/model"
	"testing"
	"time"
	usermodel "users/model"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestFormatEvents(t *testing.T) {
	userinfoval := usermodel.User{
		Id:       "abc",
		Username: "abc",
		FName:    "ab",
		LName:    "c",
		IsAdmin:  true,
		DoorAccess: map[string]bool{
			"Door1": true,
			"Door2": false,
		},
	}
	unixtime := time.Now().Unix()
	tests := []struct {
		name     string
		userinfo usermodel.User
		events   eventmodel.Events
		response model.UserResponse
	}{
		{
			name:     "Events is Empty",
			userinfo: userinfoval,
			events:   eventmodel.Events{},
			response: model.UserResponse{
				UserInfo: userinfoval,
				Events:   []map[string]string{},
			},
		},
		{
			name:     "Events found",
			userinfo: userinfoval,
			events: eventmodel.Events{
				Username: "abc",
				Events: []map[string]int64{
					map[string]int64{
						"Door1": unixtime,
					},
				},
			},
			response: model.UserResponse{
				UserInfo: userinfoval,
				Events: []map[string]string{
					map[string]string{
						"Door1": time.Unix(unixtime, 0).String(),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := FormatEvents(test.userinfo, test.events)
			if diff := cmp.Diff(actual, test.response, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("differs: (-got +want)\n%s", diff)
			}
		})
	}
}
