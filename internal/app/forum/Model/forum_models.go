package Model

import (
	"database/sql"
	"encoding/json"
	"github.com/jackc/pgtype"
)

type ForumProj struct {
	Post   int64  `json:"posts"`
	Slug   string `json:"slug"`
	Thread int32  `json:"threads"`
	Title  string `json:"title"`
	User   string `json:"user"`
}

type ThreadForum struct {
	Author  string     `json:"author"`
	Created string     `json:"created"`
	Forum   string     `json:"forum"`
	Id      int32      `json:"id"`
	Message string     `json:"message"`
	Slug    NullString `json:"slug"`
	Title   string     `json:"title"`
	Votes   int32      `json:"votes"`
}

type ThreadAns struct {
	Author  string `json:"author"`
	Created string `json:"created"`
	Forum   string `json:"forum"`
	Id      int32  `json:"id"`
	Message string `json:"message"`
	Title   string `json:"title"`
	Slug    string `json:"slug"`
	Votes   int32  `json:"votes"`
}

type NullInt64 struct {
	sql.NullInt64
}

func (v NullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullInt64) UnmarshalJSON(data []byte) error {
	var x *int64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

type NullString struct {
	sql.NullString
}

func (v NullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullString) UnmarshalJSON(data []byte) error {
	var x *string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.String = *x
	} else {
		v.Valid = false
	}
	return nil
}

type PostThread struct {
	Author   string           `json:"author"`
	Created  string           `json:"created"`
	Forum    string           `json:"forum"`
	Id       int64            `json:"id"`
	IsEdited bool             `json:"isEdited"`
	Message  string           `json:"message"`
	Parent   NullInt64        `json:"parent"`
	Thread   int32            `json:"thread"`
	Path     pgtype.Int8Array `json:"-"`
}

type VotePost struct {
	Nickname string `json:"nickname"`
	Voice    int32  `json:"voice"`
	IdThread int64  `json:"-"`
}
