update chii_tag_neue_index
set tag_results = (select count(1)
                   from chii_tag_neue_list AS tl1
                   where tl1.tlt_cat = ?
                     AND tl1.tlt_tid = chii_tag_neue_index.tag_id
                     AND tl1.tlt_type = chii_tag_neue_index.tag_type)
where tag_cat = ?
  AND tag_id IN (select distinct tl2.tlt_tid
                 from chii_tag_neue_list as tl2
                          inner join chii_tag_neue_index as ti on ti.tag_id = tl2.tlt_tid
                 where tl2.tlt_cat = ?
                   and tl2.tlt_mid = ?)
