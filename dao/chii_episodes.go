package dto

type ChiiEpisodes struct {
	EpID        int     `db:"ep_id"`
	EpSubjectID int     `db:"ep_subject_id"`
	EpSort      float64 `db:"ep_sort"`
	EpType      int     `db:"ep_type"`
	EpDisc      int     `db:"ep_disc"`
	EpName      string  `db:"ep_name"`
	EpNameCn    string  `db:"ep_name_cn"`
	EpRate      int     `db:"ep_rate"`
	EpDuration  string  `db:"ep_duration"`
	EpAirdate   string  `db:"ep_airdate"`
	EpOnline    string  `db:"ep_online"`
	EpComment   int     `db:"ep_comment"`
	EpResources int     `db:"ep_resources"`
	EpDesc      string  `db:"ep_desc"`
	EpDateline  int     `db:"ep_dateline"`
	EpLastpost  int     `db:"ep_lastpost"`
	EpLock      int     `db:"ep_lock"`
	EpBan       int     `db:"ep_ban"`
}
