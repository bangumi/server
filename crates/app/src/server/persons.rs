use async_trait::async_trait;
use axum::{
  extract::{Extension, Path, Query, State},
  http::StatusCode,
  response::Redirect,
  Json,
};
use serde::{Deserialize, Serialize};
use sqlx::QueryBuilder;
use std::collections::HashMap;
use utoipa::ToSchema;

#[cfg(test)]
use mockall::automock;

use super::media::{
  person_image, select_person_image_url, select_subject_image_url, PersonImages,
  DEFAULT_IMAGE_URL,
};
use super::{
  character_staff_string, execute_search, join_filter, parse_page, quote_str,
  search_total, staff_string, user_id_from_auth, ApiResult, AppState, MySqlExecutor,
  PageInfo, PageQuery, RequestAuth,
};

#[derive(Debug, Deserialize, Default, ToSchema)]
pub(super) struct PersonReq {
  keyword: String,
  #[serde(default)]
  filter: PersonFilter,
}

#[derive(Debug, Deserialize, Default, ToSchema)]
pub(super) struct PersonFilter {
  #[serde(default)]
  career: Vec<String>,
}

#[derive(Debug, Deserialize, Serialize)]
struct SearchHit {
  id: u32,
}

#[derive(Debug, Deserialize, Serialize, ToSchema)]
pub(super) struct PersonDoc {
  id: u32,
  name: String,
  #[serde(rename = "type")]
  person_type: u8,
  career: Vec<String>,
  short_summary: String,
  locked: bool,
  #[serde(default)]
  images: PersonImages,
}

#[derive(Debug, Serialize, ToSchema)]
pub(super) struct PersonSearchResponse {
  #[serde(flatten)]
  page: PageInfo,
  data: Vec<PersonDoc>,
}

#[derive(sqlx::FromRow)]
struct PersonRow {
  prsn_id: u32,
  prsn_name: String,
  prsn_type: u8,
  prsn_producer: bool,
  prsn_mangaka: bool,
  prsn_artist: bool,
  prsn_seiyu: bool,
  prsn_writer: bool,
  prsn_illustrator: bool,
  prsn_actor: bool,
  prsn_summary: String,
  prsn_img: String,
  prsn_lock: i8,
}

#[derive(Debug, Serialize, ToSchema)]
pub(super) struct PersonRelatedSubject {
  id: u32,
  #[serde(rename = "type")]
  subject_type: u8,
  staff: String,
  eps: String,
  name: String,
  name_cn: String,
  image: String,
}

#[derive(Debug, Serialize, ToSchema)]
pub(super) struct PersonRelatedCharacter {
  id: u32,
  name: String,
  #[serde(rename = "type")]
  character_type: u8,
  images: PersonImages,
  subject_id: u32,
  subject_type: u8,
  subject_name: String,
  subject_name_cn: String,
  staff: String,
}

#[derive(sqlx::FromRow)]
struct RelatedSubjectRow {
  subject_id: u32,
  subject_type: u8,
  subject_name: String,
  subject_name_cn: String,
  subject_image: String,
  staff_type: u16,
  eps: String,
}

#[derive(sqlx::FromRow)]
struct RelatedCharacterRow {
  character_id: u32,
  character_name: String,
  character_type: u8,
  character_img: String,
  subject_id: u32,
  subject_type: u8,
  subject_name: String,
  subject_name_cn: String,
  relation_type: u8,
}

#[derive(Debug, Deserialize, ToSchema)]
pub(super) struct ImageQuery {
  #[serde(rename = "type")]
  image_type: String,
}

#[cfg_attr(test, automock)]
#[async_trait]
pub(super) trait PersonImageRepo: Send + Sync {
  async fn find_person_image_path(
    &self,
    person_id: u32,
  ) -> Result<Option<String>, super::ApiError>;
}

struct DbPersonImageRepo<'a> {
  pool: &'a sqlx::MySqlPool,
}

#[async_trait]
impl PersonImageRepo for DbPersonImageRepo<'_> {
  async fn find_person_image_path(
    &self,
    person_id: u32,
  ) -> Result<Option<String>, super::ApiError> {
    sqlx::query_scalar::<_, String>(
      "SELECT prsn_img FROM chii_persons WHERE prsn_redirect = 0 AND prsn_id = ? LIMIT 1",
    )
    .bind(person_id)
    .fetch_optional(self.pool)
    .await
    .map_err(|_| super::ApiError::internal("load person image failed"))
  }
}

#[utoipa::path(
  post,
  path = "/v0/search/persons",
  request_body = PersonReq,
  params(PageQuery),
  responses(
    (status = 200, description = "返回搜索结果", body = PersonSearchResponse),
    (status = 400, description = "请求参数错误", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn search_persons(
  State(state): State<AppState>,
  page: Query<PageQuery>,
  Json(body): Json<PersonReq>,
) -> ApiResult<PersonSearchResponse> {
  let (limit, offset) = parse_page(page);

  let mut filters = Vec::new();
  for career in &body.filter.career {
    filters.push(format!("career = {}", quote_str(career)));
  }

  let docs = execute_search::<SearchHit>(
    &state,
    "persons",
    &body.keyword,
    limit,
    offset,
    join_filter(&filters),
    None,
  )
  .await?;

  let ids: Vec<u32> = docs.hits.iter().map(|x| x.result.id).collect();
  let data = load_persons(&state, &ids).await?;

  let total = search_total(&docs);

  Ok(Json(PersonSearchResponse {
    page: PageInfo::new(total, limit, offset),
    data,
  }))
}

#[utoipa::path(
  get,
  path = "/v0/persons/{person_id}",
  params(("person_id" = u32, Path, description = "人物 ID")),
  responses(
    (status = 200, description = "人物详情", body = PersonDoc),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_person(
  State(state): State<AppState>,
  Path(person_id): Path<u32>,
) -> ApiResult<PersonDoc> {
  let row = sqlx::query_as::<_, PersonRow>(
    "SELECT prsn_id, prsn_name, prsn_type, \
            prsn_producer, prsn_mangaka, prsn_artist, prsn_seiyu, prsn_writer, prsn_illustrator, prsn_actor, \
            prsn_summary, prsn_img, prsn_lock \
     FROM chii_persons WHERE prsn_redirect = 0 AND prsn_id = ? LIMIT 1",
  )
  .bind(person_id)
  .fetch_optional(&state.pool)
  .await
  .map_err(|_| super::ApiError::internal("load person failed"))?;

  let row = row.ok_or_else(|| super::ApiError::not_found("person not found"))?;
  Ok(Json(person_from_row(&row)))
}

#[utoipa::path(
  get,
  path = "/v0/persons/{person_id}/image",
  params(
    ("person_id" = u32, Path, description = "人物 ID"),
    ("type" = String, Query, description = "图片尺寸，可选值：small, grid, large, medium")
  ),
  responses(
    (status = 302, description = "重定向到图片地址"),
    (status = 400, description = "请求参数错误", body = super::ErrorBody),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_person_image(
  State(state): State<AppState>,
  Path(person_id): Path<u32>,
  Query(query): Query<ImageQuery>,
) -> Result<Redirect, super::ApiError> {
  let repo = DbPersonImageRepo { pool: &state.pool };
  let image_url = resolve_person_image_url(&repo, person_id, &query.image_type).await?;
  Ok(Redirect::temporary(&image_url))
}

#[utoipa::path(
  post,
  path = "/v0/persons/{person_id}/collect",
  params(("person_id" = u32, Path, description = "人物 ID")),
  responses(
    (status = 204, description = "收藏成功"),
    (status = 401, description = "未授权", body = super::ErrorBody),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn collect_person(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path(person_id): Path<u32>,
) -> Result<StatusCode, super::ApiError> {
  let user_id = user_id_from_auth(auth)?;
  collect_person_with_pool(&state.pool, user_id, person_id).await?;
  Ok(StatusCode::NO_CONTENT)
}

#[utoipa::path(
  delete,
  path = "/v0/persons/{person_id}/collect",
  params(("person_id" = u32, Path, description = "人物 ID")),
  responses(
    (status = 204, description = "取消收藏成功"),
    (status = 401, description = "未授权", body = super::ErrorBody),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn uncollect_person(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path(person_id): Path<u32>,
) -> Result<StatusCode, super::ApiError> {
  let user_id = user_id_from_auth(auth)?;
  uncollect_person_with_pool(&state.pool, user_id, person_id).await?;
  Ok(StatusCode::NO_CONTENT)
}

#[utoipa::path(
  get,
  path = "/v0/persons/{person_id}/subjects",
  params(("person_id" = u32, Path, description = "人物 ID")),
  responses(
    (status = 200, description = "人物关联条目", body = Vec<PersonRelatedSubject>),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_person_related_subjects(
  State(state): State<AppState>,
  Path(person_id): Path<u32>,
) -> ApiResult<Vec<PersonRelatedSubject>> {
  ensure_person_exists(&state.pool, person_id).await?;

  let rows: Vec<RelatedSubjectRow> = sqlx::query_as(
    "SELECT s.subject_id, s.subject_type_id AS subject_type, s.subject_name, s.subject_name_cn, s.subject_image, \
            i.prsn_position AS staff_type, i.prsn_appear_eps AS eps \
     FROM chii_person_cs_index i \
     JOIN chii_subjects s ON s.subject_id = i.subject_id \
     JOIN chii_subject_fields f ON f.field_sid = s.subject_id \
     WHERE i.prsn_type = 'prsn' AND i.prsn_id = ? AND s.subject_ban = 0 AND f.field_redirect = 0 \
     ORDER BY i.prsn_position, s.subject_id",
  )
  .bind(person_id)
  .fetch_all(&state.pool)
  .await
  .map_err(|_| super::ApiError::internal("load related subjects failed"))?;

  let data = rows
    .into_iter()
    .map(|row| PersonRelatedSubject {
      id: row.subject_id,
      subject_type: row.subject_type,
      staff: staff_string(row.subject_type, row.staff_type),
      eps: row.eps,
      name: row.subject_name,
      name_cn: row.subject_name_cn,
      image: select_subject_image_url(&row.subject_image, "large").unwrap_or_default(),
    })
    .collect();

  Ok(Json(data))
}

#[utoipa::path(
  get,
  path = "/v0/persons/{person_id}/characters",
  params(("person_id" = u32, Path, description = "人物 ID")),
  responses(
    (status = 200, description = "人物关联角色", body = Vec<PersonRelatedCharacter>),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_person_related_characters(
  State(state): State<AppState>,
  Path(person_id): Path<u32>,
) -> ApiResult<Vec<PersonRelatedCharacter>> {
  ensure_person_exists(&state.pool, person_id).await?;

  let rows: Vec<RelatedCharacterRow> = sqlx::query_as(
    "SELECT c.crt_id AS character_id, c.crt_name AS character_name, c.crt_role AS character_type, c.crt_img AS character_img, \
            s.subject_id, s.subject_type_id AS subject_type, s.subject_name, s.subject_name_cn, \
            COALESCE(si.crt_type, 0) AS relation_type \
     FROM chii_crt_cast_index ci \
     JOIN chii_characters c ON c.crt_id = ci.crt_id \
     JOIN chii_subjects s ON s.subject_id = ci.subject_id \
     JOIN chii_subject_fields f ON f.field_sid = s.subject_id \
     LEFT JOIN chii_crt_subject_index si ON si.crt_id = ci.crt_id AND si.subject_id = ci.subject_id \
     WHERE ci.prsn_id = ? AND c.crt_redirect = 0 AND s.subject_ban = 0 AND f.field_redirect = 0 \
     ORDER BY s.subject_id, c.crt_id",
  )
  .bind(person_id)
  .fetch_all(&state.pool)
  .await
  .map_err(|_| super::ApiError::internal("load related characters failed"))?;

  let data = rows
    .into_iter()
    .map(|row| PersonRelatedCharacter {
      id: row.character_id,
      name: row.character_name,
      character_type: row.character_type,
      images: person_image(&row.character_img),
      subject_id: row.subject_id,
      subject_type: row.subject_type,
      subject_name: row.subject_name,
      subject_name_cn: row.subject_name_cn,
      staff: character_staff_string(row.relation_type),
    })
    .collect();

  Ok(Json(data))
}

async fn ensure_person_exists<'e>(
  executor: impl MySqlExecutor<'e>,
  person_id: u32,
) -> Result<(), super::ApiError> {
  let exists = sqlx::query_scalar::<_, u32>(
    "SELECT prsn_id FROM chii_persons WHERE prsn_redirect = 0 AND prsn_id = ? LIMIT 1",
  )
  .bind(person_id)
  .fetch_optional(executor)
  .await
  .map_err(|_| super::ApiError::internal("load person failed"))?
  .is_some();

  if !exists {
    return Err(super::ApiError::not_found("person not found"));
  }

  Ok(())
}

pub(super) async fn collect_person_with_pool(
  pool: &sqlx::MySqlPool,
  user_id: u32,
  person_id: u32,
) -> Result<(), super::ApiError> {
  ensure_person_exists(pool, person_id).await?;

  let exists = sqlx::query_scalar::<_, u32>(
    "SELECT prsn_clt_id FROM chii_person_collects WHERE prsn_clt_cat = 'prsn' AND prsn_clt_uid = ? AND prsn_clt_mid = ? LIMIT 1",
  )
  .bind(user_id)
  .bind(person_id)
  .fetch_optional(pool)
  .await
  .map_err(|_| super::ApiError::internal("query person collect failed"))?
  .is_some();

  if exists {
    return Ok(());
  }

  sqlx::query(
    "INSERT INTO chii_person_collects (prsn_clt_cat, prsn_clt_mid, prsn_clt_uid, prsn_clt_dateline) VALUES ('prsn', ?, ?, UNIX_TIMESTAMP())",
  )
  .bind(person_id)
  .bind(user_id)
  .execute(pool)
  .await
  .map_err(|_| super::ApiError::internal("add person collect failed"))?;

  Ok(())
}

pub(super) async fn uncollect_person_with_pool(
  pool: &sqlx::MySqlPool,
  user_id: u32,
  person_id: u32,
) -> Result<(), super::ApiError> {
  ensure_person_exists(pool, person_id).await?;

  let result = sqlx::query(
    "DELETE FROM chii_person_collects WHERE prsn_clt_cat = 'prsn' AND prsn_clt_uid = ? AND prsn_clt_mid = ?",
  )
  .bind(user_id)
  .bind(person_id)
  .execute(pool)
  .await
  .map_err(|_| super::ApiError::internal("remove person collect failed"))?;

  if result.rows_affected() == 0 {
    return Err(super::ApiError::not_found("person not collected"));
  }

  Ok(())
}

async fn resolve_person_image_url(
  repo: &impl PersonImageRepo,
  person_id: u32,
  image_type: &str,
) -> Result<String, super::ApiError> {
  let path = repo.find_person_image_path(person_id).await?;
  let path = path.ok_or_else(|| super::ApiError::not_found("person not found"))?;

  let image_url = select_person_image_url(&path, image_type).ok_or_else(|| {
    super::ApiError::bad_request(format!("bad image type: {image_type}"))
  })?;

  if image_url.is_empty() {
    return Ok(DEFAULT_IMAGE_URL.to_string());
  }

  Ok(image_url)
}

async fn load_persons(
  state: &AppState,
  ids: &[u32],
) -> Result<Vec<PersonDoc>, super::ApiError> {
  if ids.is_empty() {
    return Ok(Vec::new());
  }

  let mut qb = QueryBuilder::new(
    "SELECT prsn_id, prsn_name, prsn_type, \
            prsn_producer, prsn_mangaka, prsn_artist, prsn_seiyu, prsn_writer, prsn_illustrator, prsn_actor, \
            prsn_summary, prsn_img, prsn_lock \
     FROM chii_persons WHERE prsn_redirect = 0 AND prsn_id IN (",
  );

  {
    let mut separated = qb.separated(", ");
    for id in ids {
      separated.push_bind(*id);
    }
  }
  qb.push(")");

  let rows: Vec<PersonRow> = qb
    .build_query_as()
    .fetch_all(&state.pool)
    .await
    .map_err(|_| super::ApiError::internal("load persons failed"))?;

  let by_id: HashMap<u32, PersonRow> =
    rows.into_iter().map(|x| (x.prsn_id, x)).collect();

  let mut result = Vec::with_capacity(ids.len());
  for id in ids {
    if let Some(row) = by_id.get(id) {
      result.push(person_from_row(row));
    }
  }

  Ok(result)
}

fn person_from_row(row: &PersonRow) -> PersonDoc {
  PersonDoc {
    id: row.prsn_id,
    name: row.prsn_name.clone(),
    person_type: row.prsn_type,
    career: careers(row),
    short_summary: row.prsn_summary.clone(),
    locked: row.prsn_lock != 0,
    images: person_image(&row.prsn_img),
  }
}

fn careers(row: &PersonRow) -> Vec<String> {
  let mut items = Vec::with_capacity(7);

  if row.prsn_writer {
    items.push("writer".to_string());
  }
  if row.prsn_producer {
    items.push("producer".to_string());
  }
  if row.prsn_mangaka {
    items.push("mangaka".to_string());
  }
  if row.prsn_artist {
    items.push("artist".to_string());
  }
  if row.prsn_seiyu {
    items.push("seiyu".to_string());
  }
  if row.prsn_illustrator {
    items.push("illustrator".to_string());
  }
  if row.prsn_actor {
    items.push("actor".to_string());
  }

  items
}

#[cfg(test)]
mod tests {
  use super::resolve_person_image_url;
  use crate::server::test_mocks::MockPool;
  use axum::http::StatusCode;

  #[tokio::test]
  async fn resolve_person_image_url_returns_default_for_empty_path() {
    let mut pool = MockPool::new();
    pool
      .person_image_repo
      .expect_find_person_image_path()
      .withf(|person_id| *person_id == 9)
      .times(1)
      .returning(|_| Ok(Some(String::new())));

    let got = resolve_person_image_url(&pool.person_image_repo, 9, "small")
      .await
      .expect("resolve image");

    assert_eq!(got, "https://lain.bgm.tv/img/no_icon_subject.png");
  }

  #[tokio::test]
  async fn resolve_person_image_url_returns_not_found_when_missing() {
    let mut pool = MockPool::new();
    pool
      .person_image_repo
      .expect_find_person_image_path()
      .withf(|person_id| *person_id == 404)
      .times(1)
      .returning(|_| Ok(None));

    let err = resolve_person_image_url(&pool.person_image_repo, 404, "small")
      .await
      .expect_err("expect not found");

    assert_eq!(err.status, StatusCode::NOT_FOUND);
    assert_eq!(err.message, "person not found");
  }

  #[tokio::test]
  async fn resolve_person_image_url_returns_bad_request_for_invalid_type() {
    let mut pool = MockPool::new();
    pool
      .person_image_repo
      .expect_find_person_image_path()
      .times(1)
      .returning(|_| Ok(Some("ab/cd.jpg".to_string())));

    let err = resolve_person_image_url(&pool.person_image_repo, 1, "invalid")
      .await
      .expect_err("expect bad request");

    assert_eq!(err.status, StatusCode::BAD_REQUEST);
    assert_eq!(err.message, "bad image type: invalid");
  }
}
