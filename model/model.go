package model

import (
	usermodel "users/model"
)

type UserResponse struct {
	UserInfo usermodel.User      `json:"userinfo"`
	Events   []map[string]string `json:"events"`
}
