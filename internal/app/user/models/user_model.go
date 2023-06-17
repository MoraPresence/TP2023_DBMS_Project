package models

type UserInfo struct {
	About string `json:"about"`
	Email string `json:"email"`
	Full  string `json:"fullname"`
	Nick  string `json:"nickname"`
}
