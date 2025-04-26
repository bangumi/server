package dto

type ChiiEpStatus struct {
	EpSttID        int    `db:"ep_stt_id"`
	EpSttUID       int    `db:"ep_stt_uid"`
	EpSttSID       int    `db:"ep_stt_sid"`
	EpSttOnPrg     int    `db:"ep_stt_on_prg"`
	EpSttStatus    string `db:"ep_stt_status"`
	EpSttLasttouch int    `db:"ep_stt_lasttouch"`
}
