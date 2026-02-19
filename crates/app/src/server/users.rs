use axum::{
  extract::{Extension, Path, Query, State},
  Json,
};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use utoipa::{IntoParams, ToSchema};

use super::media::{subject_image, SubjectImages};
use super::{
  parse_page, ApiError, ApiResult, AppState, PageInfo, PageQuery, RequestAuth,
};

#[derive(Debug, Deserialize, IntoParams)]
#[into_params(parameter_in = Query)]
pub(super) struct UserCollectionsQuery {
  subject_type: Option<u8>,
  #[serde(rename = "type")]
  collection_type: Option<u8>,
}

#[derive(Debug, Serialize, ToSchema)]
pub(super) struct UserSubjectCollection {
  subject_id: u32,
  subject_type: u8,
  rate: u8,
  #[serde(rename = "type")]
  collection_type: u8,
  comment: Option<String>,
  tags: Vec<String>,
  ep_status: u32,
  vol_status: u32,
  updated_at: String,
  private: bool,
  subject: SlimSubject,
}

#[derive(Debug, Serialize, ToSchema)]
pub(super) struct UserCollectionsResponse {
  #[serde(flatten)]
  page: PageInfo,
  data: Vec<UserSubjectCollection>,
}

#[derive(Debug, Serialize, ToSchema)]
struct SlimSubject {
  id: u32,
  #[serde(rename = "type")]
  subject_type: u8,
  name: String,
  name_cn: String,
  short_summary: String,
  date: Option<String>,
  images: SubjectImages,
  volumes: u32,
  eps: u32,
  collection_total: u32,
  score: f64,
  rank: u32,
  tags: Vec<SubjectTag>,
}

#[derive(Debug, Serialize, ToSchema)]
struct SubjectTag {
  name: String,
  count: u32,
}

#[derive(Debug, FromRow)]
struct SubjectCollectionRow {
  subject_id: u32,
  subject_type: u8,
  rate: u8,
  collection_type: u8,
  comment: String,
  tags: String,
  ep_status: u32,
  vol_status: u32,
  updated_at: String,
  private: u8,
  subject_name: String,
  subject_name_cn: String,
  short_summary: String,
  date: Option<String>,
  subject_image: String,
  volumes: u32,
  eps: u32,
  collection_total: u32,
  rank: u32,
  rate_1: u32,
  rate_2: u32,
  rate_3: u32,
  rate_4: u32,
  rate_5: u32,
  rate_6: u32,
  rate_7: u32,
  rate_8: u32,
  rate_9: u32,
  rate_10: u32,
  field_tags: Vec<u8>,
}

#[utoipa::path(
  get,
  path = "/v0/users/{username}/collections",
  params(
    ("username" = String, Path, description = "用户名"),
    UserCollectionsQuery,
    PageQuery,
  ),
  responses(
    (status = 200, description = "用户收藏列表", body = UserCollectionsResponse),
    (status = 400, description = "请求参数错误", body = super::ErrorBody),
    (status = 404, description = "用户不存在", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn list_user_collections(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path(username): Path<String>,
  query: Query<UserCollectionsQuery>,
  page: Query<PageQuery>,
) -> ApiResult<UserCollectionsResponse> {
  let user_id = find_user_id_by_username(&state, &username).await?;

  let (subject_type, collection_type) = parse_collection_filters(&query.0)?;
  let (limit, offset) = parse_page(page);
  let show_private = auth.user_id == Some(user_id);

  let total = count_subject_collections(
    &state,
    user_id,
    subject_type,
    collection_type,
    show_private,
  )
  .await?;

  let rows = list_subject_collections(
    &state,
    user_id,
    subject_type,
    collection_type,
    show_private,
    limit,
    offset,
  )
  .await?;

  Ok(Json(UserCollectionsResponse {
    page: PageInfo::new(total as usize, limit, offset),
    data: rows.into_iter().map(map_row_to_collection).collect(),
  }))
}

#[utoipa::path(
  get,
  path = "/v0/users/{username}/collections/{subject_id}",
  params(
    ("username" = String, Path, description = "用户名"),
    ("subject_id" = u32, Path, description = "条目 ID")
  ),
  responses(
    (status = 200, description = "用户收藏", body = UserSubjectCollection),
    (status = 400, description = "请求参数错误", body = super::ErrorBody),
    (status = 404, description = "用户不存在或未收藏", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_user_collection(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path((username, subject_id)): Path<(String, u32)>,
) -> ApiResult<UserSubjectCollection> {
  let user_id = find_user_id_by_username(&state, &username).await?;
  let show_private = auth.user_id == Some(user_id);

  let row =
    get_subject_collection(&state, user_id, subject_id, show_private, auth.allow_nsfw)
      .await?;
  Ok(Json(map_row_to_collection(row)))
}

async fn find_user_id_by_username(
  state: &AppState,
  username: &str,
) -> Result<u32, ApiError> {
  sqlx::query_scalar::<_, u32>(
    "SELECT uid FROM chii_members WHERE username = ? LIMIT 1",
  )
  .bind(username)
  .fetch_optional(state.pool())
  .await
  .map_err(|_| ApiError::internal("load user failed"))?
  .ok_or_else(|| ApiError::not_found("user doesn't exist or has been removed"))
}

fn parse_collection_filters(
  query: &UserCollectionsQuery,
) -> Result<(Option<u8>, Option<u8>), ApiError> {
  if let Some(subject_type) = query.subject_type {
    if !matches!(subject_type, 1 | 2 | 3 | 4 | 6) {
      return Err(ApiError::bad_request("invalid query param `subject_type`"));
    }
  }

  if let Some(collection_type) = query.collection_type {
    if !(1..=5).contains(&collection_type) {
      return Err(ApiError::bad_request("invalid query param `type`"));
    }
  }

  Ok((query.subject_type, query.collection_type))
}

async fn count_subject_collections(
  state: &AppState,
  user_id: u32,
  subject_type: Option<u8>,
  collection_type: Option<u8>,
  show_private: bool,
) -> Result<i64, ApiError> {
  let mut sql = String::from(
    "SELECT COUNT(*) AS cnt FROM chii_subject_interests i WHERE i.interest_uid = ? AND i.interest_type != 0",
  );

  if subject_type.is_some() {
    sql.push_str(" AND i.interest_subject_type = ?");
  }
  if collection_type.is_some() {
    sql.push_str(" AND i.interest_type = ?");
  }
  if !show_private {
    sql.push_str(" AND i.interest_private = 0");
  }

  let mut q = sqlx::query_scalar::<_, i64>(&sql).bind(user_id);
  if let Some(v) = subject_type {
    q = q.bind(v);
  }
  if let Some(v) = collection_type {
    q = q.bind(v);
  }

  q.fetch_one(state.pool())
    .await
    .map_err(|_| ApiError::internal("count user collections failed"))
}

async fn list_subject_collections(
  state: &AppState,
  user_id: u32,
  subject_type: Option<u8>,
  collection_type: Option<u8>,
  show_private: bool,
  limit: usize,
  offset: usize,
) -> Result<Vec<SubjectCollectionRow>, ApiError> {
  let mut sql = String::from(
    "SELECT i.interest_subject_id AS subject_id, i.interest_subject_type AS subject_type, i.interest_rate AS rate, \
            i.interest_type AS collection_type, i.interest_comment AS comment, i.interest_tag AS tags, \
            i.interest_ep_status AS ep_status, i.interest_vol_status AS vol_status, \
            DATE_FORMAT(FROM_UNIXTIME(i.interest_lasttouch), '%Y-%m-%dT%H:%i:%s+08:00') AS updated_at, \
            i.interest_private AS private, \
            s.subject_name, s.subject_name_cn, s.field_summary AS short_summary, DATE_FORMAT(f.field_date, '%Y-%m-%d') AS date, \
            s.subject_image, s.field_volumes AS volumes, s.field_eps AS eps, s.subject_collect AS collection_total, \
            f.field_rank AS rank, \
            f.field_rate_1 AS rate_1, f.field_rate_2 AS rate_2, f.field_rate_3 AS rate_3, f.field_rate_4 AS rate_4, f.field_rate_5 AS rate_5, \
            f.field_rate_6 AS rate_6, f.field_rate_7 AS rate_7, f.field_rate_8 AS rate_8, f.field_rate_9 AS rate_9, f.field_rate_10 AS rate_10, \
            f.field_tags \
     FROM chii_subject_interests i \
     JOIN chii_subjects s ON s.subject_id = i.interest_subject_id \
     JOIN chii_subject_fields f ON f.field_sid = s.subject_id \
     WHERE i.interest_uid = ? AND i.interest_type != 0 AND s.subject_ban = 0 AND f.field_redirect = 0",
  );

  if subject_type.is_some() {
    sql.push_str(" AND i.interest_subject_type = ?");
  }
  if collection_type.is_some() {
    sql.push_str(" AND i.interest_type = ?");
  }
  if !show_private {
    sql.push_str(" AND i.interest_private = 0");
  }

  sql.push_str(" ORDER BY i.interest_lasttouch DESC LIMIT ? OFFSET ?");

  let mut q = sqlx::query_as::<_, SubjectCollectionRow>(&sql).bind(user_id);
  if let Some(v) = subject_type {
    q = q.bind(v);
  }
  if let Some(v) = collection_type {
    q = q.bind(v);
  }

  q.bind(limit as i64)
    .bind(offset as i64)
    .fetch_all(state.pool())
    .await
    .map_err(|_| ApiError::internal("list user collections failed"))
}

async fn get_subject_collection(
  state: &AppState,
  user_id: u32,
  subject_id: u32,
  show_private: bool,
  allow_nsfw: bool,
) -> Result<SubjectCollectionRow, ApiError> {
  let mut sql = String::from(
    "SELECT i.interest_subject_id AS subject_id, i.interest_subject_type AS subject_type, i.interest_rate AS rate, \
            i.interest_type AS collection_type, i.interest_comment AS comment, i.interest_tag AS tags, \
            i.interest_ep_status AS ep_status, i.interest_vol_status AS vol_status, \
            DATE_FORMAT(FROM_UNIXTIME(i.interest_lasttouch), '%Y-%m-%dT%H:%i:%s+08:00') AS updated_at, \
            i.interest_private AS private, \
            s.subject_name, s.subject_name_cn, s.field_summary AS short_summary, DATE_FORMAT(f.field_date, '%Y-%m-%d') AS date, \
            s.subject_image, s.field_volumes AS volumes, s.field_eps AS eps, s.subject_collect AS collection_total, \
            f.field_rank AS rank, \
            f.field_rate_1 AS rate_1, f.field_rate_2 AS rate_2, f.field_rate_3 AS rate_3, f.field_rate_4 AS rate_4, f.field_rate_5 AS rate_5, \
            f.field_rate_6 AS rate_6, f.field_rate_7 AS rate_7, f.field_rate_8 AS rate_8, f.field_rate_9 AS rate_9, f.field_rate_10 AS rate_10, \
            f.field_tags \
     FROM chii_subject_interests i \
     JOIN chii_subjects s ON s.subject_id = i.interest_subject_id \
     JOIN chii_subject_fields f ON f.field_sid = s.subject_id \
     WHERE i.interest_uid = ? AND i.interest_subject_id = ? AND i.interest_type != 0 \
       AND s.subject_ban = 0 AND f.field_redirect = 0",
  );

  if !allow_nsfw {
    sql.push_str(" AND s.subject_nsfw = 0");
  }

  let row = sqlx::query_as::<_, SubjectCollectionRow>(&sql)
    .bind(user_id)
    .bind(subject_id)
    .fetch_optional(state.pool())
    .await
    .map_err(|_| ApiError::internal("load user collection failed"))?
    .ok_or_else(|| ApiError::not_found("subject is not collected by user"))?;

  if row.private != 0 && !show_private {
    return Err(ApiError::not_found("subject is not collected by user"));
  }

  Ok(row)
}

fn map_row_to_collection(row: SubjectCollectionRow) -> UserSubjectCollection {
  let score = rating_score(&row);

  UserSubjectCollection {
    subject_id: row.subject_id,
    subject_type: row.subject_type,
    rate: row.rate,
    collection_type: row.collection_type,
    comment: if row.comment.is_empty() {
      None
    } else {
      Some(row.comment)
    },
    tags: split_tags(&row.tags),
    ep_status: row.ep_status,
    vol_status: row.vol_status,
    updated_at: row.updated_at,
    private: row.private != 0,
    subject: SlimSubject {
      id: row.subject_id,
      subject_type: row.subject_type,
      name: row.subject_name,
      name_cn: row.subject_name_cn,
      short_summary: row.short_summary,
      date: row.date,
      images: subject_image(&row.subject_image),
      volumes: row.volumes,
      eps: row.eps,
      collection_total: row.collection_total,
      score,
      rank: row.rank,
      tags: parse_subject_tags(&row.field_tags),
    },
  }
}

fn split_tags(raw: &str) -> Vec<String> {
  raw
    .split(' ')
    .map(str::trim)
    .filter(|x| !x.is_empty())
    .map(ToOwned::to_owned)
    .collect()
}

fn parse_subject_tags(raw: &[u8]) -> Vec<SubjectTag> {
  let s = String::from_utf8_lossy(raw);
  let parsed: Vec<SubjectTagItem> = match php_serialize::from_str(&s) {
    Ok(v) => v,
    Err(_) => return Vec::new(),
  };

  parsed
    .into_iter()
    .filter_map(|x| {
      x.tag_name.map(|name| SubjectTag {
        name,
        count: x.tag_count.unwrap_or(0),
      })
    })
    .collect()
}

fn rating_score(row: &SubjectCollectionRow) -> f64 {
  let total = row.rate_1
    + row.rate_2
    + row.rate_3
    + row.rate_4
    + row.rate_5
    + row.rate_6
    + row.rate_7
    + row.rate_8
    + row.rate_9
    + row.rate_10;

  if total == 0 {
    return 0.0;
  }

  let weighted = (row.rate_1 as f64) * 1.0
    + (row.rate_2 as f64) * 2.0
    + (row.rate_3 as f64) * 3.0
    + (row.rate_4 as f64) * 4.0
    + (row.rate_5 as f64) * 5.0
    + (row.rate_6 as f64) * 6.0
    + (row.rate_7 as f64) * 7.0
    + (row.rate_8 as f64) * 8.0
    + (row.rate_9 as f64) * 9.0
    + (row.rate_10 as f64) * 10.0;

  ((weighted / (total as f64)) * 10.0).round() / 10.0
}

#[derive(Debug, Deserialize)]
struct SubjectTagItem {
  tag_name: Option<String>,
  tag_count: Option<u32>,
}
