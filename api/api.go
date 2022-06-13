package api

import (
	"accessdoor/model"
	eventmodel "events/model"
	"time"
	usermodel "users/model"
)

func FormatEvents(usrinfo usermodel.User, events eventmodel.Events) model.UserResponse {
	formattedevent := make([]map[string]string, len(events.Events))
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
