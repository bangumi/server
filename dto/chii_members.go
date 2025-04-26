package dto

import (
	"time"
)

type ChiiMembers struct {
	UID             int       `db:"uid"`
	Username        string    `db:"username"`
	Nickname        string    `db:"nickname"`
	Avatar          string    `db:"avatar"`
	Secques         string    `db:"secques"`
	Gender          int       `db:"gender"`
	AdminID         int       `db:"adminid"`
	GroupID         int       `db:"groupid"`
	Regip           string    `db:"regip"`
	Regdate         int       `db:"regdate"`
	Lastip          string    `db:"lastip"`
	Lastvisit       int       `db:"lastvisit"`
	Lastactivity    int       `db:"lastactivity"`
	Lastpost        int       `db:"lastpost"`
	Bday            time.Time `db:"bday"`
	StyleID         int       `db:"styleid"`
	Dateformat      string    `db:"dateformat"`
	Timeformat      int       `db:"timeformat"`
	Newsletter      int       `db:"newsletter"`
	Timeoffset      string    `db:"timeoffset"`
	Newpm           int       `db:"newpm"`
	NewNotify       int       `db:"new_notify"`
	UsernameLock    int       `db:"username_lock"`
	UkagakaSettings string    `db:"ukagaka_settings"`
	ImgChart        int       `db:"img_chart"`
	Sign            string    `db:"sign"`
	PasswordCrypt   string    `db:"password_crypt"`
	Email           string    `db:"email"`
	Acl             string    `db:"acl"`
	Invited         int       `db:"invited"`
}
