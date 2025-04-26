package dto

type ChiiMemberfields struct {
	UID                   int     `db:"uid"`
	Site                  string  `db:"site"`
	Location              string  `db:"location"`
	Bio                   string  `db:"bio"`
	Homepage              string  `db:"homepage"`
	IndexSort             string  `db:"index_sort"`
	UserAgent             string  `db:"user_agent"`
	Ignorepm              string  `db:"ignorepm"`
	Groupterms            string  `db:"groupterms"`
	Authstr               string  `db:"authstr"`
	Privacy               string  `db:"privacy"`
	Blocklist             string  `db:"blocklist"`
	RegSource             int     `db:"reg_source"`
	InviteNum             int     `db:"invite_num"`
	EmailVerified         int     `db:"email_verified"`
	EmailVerifyToken      string  `db:"email_verify_token"`
	EmailVerifyScore      float64 `db:"email_verify_score"`
	EmailVerifyDateline   int     `db:"email_verify_dateline"`
	ResetPasswordForce    int     `db:"reset_password_force"`
	ResetPasswordToken    string  `db:"reset_password_token"`
	ResetPasswordDateline int     `db:"reset_password_dateline"`
}
