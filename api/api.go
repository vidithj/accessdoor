package api

import (
	"accessdoor/model"
	eventmodel "events/model"
	"time"
	usermodel "users/model"
)

func FormatEvents(usrinfo usermodel.User, events eventmodel.Events) model.UserResponse {
	if len(events.Events) == 0 {
		return model.UserResponse{
			UserInfo: usrinfo,
			Events:   []map[string]string{},
		}
	}
	formattedevent := []map[string]string{}
	for _, val := range events.Events {
		for key, unixtime := range val {
			timeStamp := time.Unix(unixtime, 0)
			formattedevent = append(formattedevent, map[string]string{
				key: timeStamp.String(),
			})
		}
	}
	return model.UserResponse{
		UserInfo: usrinfo,
		Events:   formattedevent,
	}
}
