package dto

type ChiiApps struct {
	AppID           int    `db:"app_id"`
	AppType         int    `db:"app_type"`
	AppCreator      int    `db:"app_creator"`
	AppName         string `db:"app_name"`
	AppDesc         string `db:"app_desc"`
	AppURL          string `db:"app_url"`
	AppIsThirdParty int    `db:"app_is_third_party"`
	AppCollects     int    `db:"app_collects"`
	AppStatus       int    `db:"app_status"`
	AppTimestamp    int    `db:"app_timestamp"`
	AppLasttouch    int    `db:"app_lasttouch"`
	AppBan          int    `db:"app_ban"`
}
