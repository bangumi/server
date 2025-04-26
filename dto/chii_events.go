package dto

type ChiiEvents struct {
	EventID         int    `db:"event_id"`
	EventTitle      string `db:"event_title"`
	EventType       int    `db:"event_type"`
	EventCreator    int    `db:"event_creator"`
	EventStartTime  int    `db:"event_start_time"`
	EventEndTime    int    `db:"event_end_time"`
	EventImg        string `db:"event_img"`
	EventState      int    `db:"event_state"`
	EventCity       int    `db:"event_city"`
	EventAddress    string `db:"event_address"`
	EventDesc       string `db:"event_desc"`
	EventWish       int    `db:"event_wish"`
	EventDo         int    `db:"event_do"`
	EventBuildtime  int    `db:"event_buildtime"`
	EventLastupdate int    `db:"event_lastupdate"`
}
