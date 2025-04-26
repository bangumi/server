package dto

type ChiiGroups struct {
	GrpID         int    `db:"grp_id"`
	GrpCat        int    `db:"grp_cat"`
	GrpName       string `db:"grp_name"`
	GrpTitle      string `db:"grp_title"`
	GrpIcon       string `db:"grp_icon"`
	GrpCreator    int    `db:"grp_creator"`
	GrpTopics     int    `db:"grp_topics"`
	GrpPosts      int    `db:"grp_posts"`
	GrpMembers    int    `db:"grp_members"`
	GrpDesc       string `db:"grp_desc"`
	GrpLastpost   int    `db:"grp_lastpost"`
	GrpBuilddate  int    `db:"grp_builddate"`
	GrpAccessible int    `db:"grp_accessible"`
	GrpNsfw       int    `db:"grp_nsfw"`
}
