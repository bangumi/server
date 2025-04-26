package dto

type ChiiEden struct {
	EdenID            int    `db:"eden_id"`
	EdenType          int    `db:"eden_type"`
	EdenName          string `db:"eden_name"`
	EdenTitle         string `db:"eden_title"`
	EdenIcon          string `db:"eden_icon"`
	EdenHeader        string `db:"eden_header"`
	EdenDesc          string `db:"eden_desc"`
	EdenRelateSubject string `db:"eden_relate_subject"`
	EdenRelateGrp     string `db:"eden_relate_grp"`
	EdenMembers       int    `db:"eden_members"`
	EdenLasttouch     int    `db:"eden_lasttouch"`
	EdenBuilddate     int    `db:"eden_builddate"`
	EdenPushdate      int    `db:"eden_pushdate"`
}
