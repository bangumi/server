<?php
class TimelineCore
{
    const TYPE_ALL          = 0;
    const TYPE_DAILY        = 1;
    const TYPE_WIKI         = 2;
    const TYPE_SBJ_COLLECT  = 3;
    const TYPE_BGM_PROGRESS = 4;
    const TYPE_BGM_PROGRESS_BATCH = 0;
    const TYPE_BGM_PROGRESS_SINGLE = 1;

    const TYPE_STATUS       = 5;
    const TYPE_BLOG         = 6;
    const TYPE_INDEX        = 7;
    const TYPE_MONO         = 8;
    const TYPE_DOUJIN       = 9;

    const TYPE_STATUS_SIGN     = 0;
    const TYPE_STATUS_TSUKKOMI = 1;
    const TYPE_STATUS_NICKNAME = 2;

    const CAT_DOUJIN_SBJ_NEW      = 0;
    const CAT_DOUJIN_SBJ_COLLECT  = 1;
    const CAT_DOUJIN_CLUB_NEW     = 2;
    const CAT_DOUJIN_CLUB_FOLLOW  = 3;
    const CAT_DOUJIN_EVENT_FOLLOW = 4;
    const CAT_DOUJIN_EVENT_COLLECT = 5;
    const CAT_DOUJIN_EVENT_NEW    = 6;

    const NOTIFY_STATUS_ALL = -1;
    const NOTIFY_STATUS_UNREAD = 1;
    const NOTIFY_STATUS_READ = 0;

    /**
     *
     *
     * @param unknown $type (optional)
     * @return unknown
     */
    public static function TimelineTypeConv($type = 'all')
    {
        $info = array();
        switch ($type) {
            default:
            case 'all':
                $info['id'] = '0';
                $info['url'] = 'all';
                $info['name'] = '全部';
                break;
            case 'subject':
                $info['id'] = '1';
                $info['url'] = 'subject';
                $info['name'] = '收藏';
                $info['sql'] = 'tml_cat = 3';
                break;
            case 'progress':
                $info['id'] = '2';
                $info['url'] = 'progress';
                $info['name'] = '进度';
                $info['sql'] = 'tml_cat = 4';
                break;
            case 'relation':
                $info['id'] = '3';
                $info['url'] = 'relation';
                $info['name'] = '好友';
                $info['sql'] = 'tml_cat = 1 AND tml_type =2';
                break;
            case 'group':
                $info['id'] = '4';
                $info['url'] = 'group';
                $info['name'] = '小组';
                $info['sql'] = 'tml_cat = 1 AND tml_type IN (3,4)';
                break;
            case 'say':
                $info['id'] = '5';
                $info['url'] = 'say';
                $info['name'] = '吐槽';
                $info['sql'] = 'tml_cat = 5';
                break;
            case 'wiki':
                $info['id'] = '6';
                $info['url'] = 'wiki';
                $info['name'] = '维基';
                $info['sql'] = 'tml_cat = 2';
                break;
            case 'blog':
                $info['id'] = '7';
                $info['url'] = 'blog';
                $info['name'] = '日志';
                $info['sql'] = 'tml_cat = 6';
                break;
            case 'index':
                $info['id'] = '8';
                $info['url'] = 'index';
                $info['name'] = '目录';
                $info['sql'] = 'tml_cat = 7';
                break;
            case 'mono':
                $info['id'] = self::TYPE_MONO;
                $info['url'] = 'mono';
                $info['name'] = '人物';
                $info['sql'] = 'tml_cat = ' . self::TYPE_MONO;
                break;
            case 'doujin':
                $info['id'] = self::TYPE_DOUJIN;
                $info['url'] = 'doujin';
                $info['name'] = '天窗';
                $info['sql'] = 'tml_cat = ' . self::TYPE_DOUJIN;
                break;
            case 'replies':
                $info['id'] = '999';
                $info['url'] = 'replies';
                $info['name'] = '回复';
                $info['sql'] = '';
                break;
        }
        return $info;
    }
}
