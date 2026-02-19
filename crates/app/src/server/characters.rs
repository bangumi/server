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
  character_staff_string, execute_search, join_filter, parse_page, search_total,
  user_id_from_auth, ApiResult, AppState, MySqlExecutor, PageInfo, PageQuery,
  RequestAuth,
};

#[derive(Debug, Deserialize, Default, ToSchema)]
pub(super) struct CharacterReq {
  keyword: String,
  #[serde(default)]
  filter: CharacterFilter,
}

#[derive(Debug, Deserialize, Default, ToSchema)]
pub(super) struct CharacterFilter {
  nsfw: Option<bool>,
}

#[derive(Debug, Deserialize, Serialize)]
struct SearchHit {
  id: u32,
}

#[derive(Debug, Deserialize, Serialize, ToSchema)]
pub(super) struct CharacterDoc {
  id: u32,
  name: String,
  #[serde(rename = "type")]
  role: u8,
  summary: String,
  locked: bool,
  #[serde(default)]
  images: PersonImages,
  stat: Stat,
  gender: Option<String>,
  blood_type: Option<u8>,
  birth_year: Option<u16>,
  birth_mon: Option<u8>,
  birth_day: Option<u8>,
}

#[derive(Debug, Deserialize, Serialize, ToSchema)]
pub(super) struct Stat {
  comments: u32,
  collects: u32,
}

#[derive(Debug, Serialize, ToSchema)]
pub(super) struct CharacterSearchResponse {
  #[serde(flatten)]
  page: PageInfo,
  data: Vec<CharacterDoc>,
}

#[derive(sqlx::FromRow)]
struct CharacterRow {
  crt_id: u32,
  crt_name: String,
  crt_role: u8,
  crt_summary: String,
  crt_img: String,
  crt_comment: u32,
  crt_collects: u32,
  crt_lock: i8,
  gender: Option<u8>,
  bloodtype: Option<u8>,
  birth_year: Option<u16>,
  birth_mon: Option<u8>,
  birth_day: Option<u8>,
}

#[derive(Debug, Serialize, ToSchema)]
pub(super) struct CharacterRelatedSubject {
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
pub(super) struct CharacterRelatedPerson {
  id: u32,
  name: String,
  #[serde(rename = "type")]
  person_type: u8,
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
  relation_type: u8,
  eps: String,
}

#[derive(sqlx::FromRow)]
struct RelatedPersonRow {
  person_id: u32,
  person_name: String,
  person_type: u8,
  person_img: String,
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
pub(super) trait CharacterImageRepo: Send + Sync {
  async fn find_character_image_path(
    &self,
    character_id: u32,
  ) -> Result<Option<String>, super::ApiError>;
}

struct DbCharacterImageRepo<'a> {
  pool: &'a sqlx::MySqlPool,
}

#[async_trait]
impl CharacterImageRepo for DbCharacterImageRepo<'_> {
  async fn find_character_image_path(
    &self,
    character_id: u32,
  ) -> Result<Option<String>, super::ApiError> {
    sqlx::query_scalar::<_, String>(
      "SELECT crt_img FROM chii_characters WHERE crt_redirect = 0 AND crt_id = ? LIMIT 1",
    )
    .bind(character_id)
    .fetch_optional(self.pool)
    .await
    .map_err(|_| super::ApiError::internal("load character image failed"))
  }
}

#[utoipa::path(
  post,
  path = "/v0/search/characters",
  request_body = CharacterReq,
  params(PageQuery),
  responses(
    (status = 200, description = "返回搜索结果", body = CharacterSearchResponse),
    (status = 400, description = "请求参数错误", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn search_characters(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  page: Query<PageQuery>,
  Json(body): Json<CharacterReq>,
) -> ApiResult<CharacterSearchResponse> {
  let (limit, offset) = parse_page(page);
  let allow_nsfw = auth.allow_nsfw;

  let mut filters = Vec::new();
  if !allow_nsfw {
    filters.push("nsfw = false".to_string());
  } else if let Some(v) = body.filter.nsfw {
    filters.push(format!("nsfw = {v}"));
  }

  let docs = execute_search::<SearchHit>(
    &state,
    "characters",
    &body.keyword,
    limit,
    offset,
    join_filter(&filters),
    None,
  )
  .await?;

  let ids: Vec<u32> = docs.hits.iter().map(|x| x.result.id).collect();
  let data = load_characters(&state, &ids).await?;

  let total = search_total(&docs);

  Ok(Json(CharacterSearchResponse {
    page: PageInfo::new(total, limit, offset),
    data,
  }))
}

#[utoipa::path(
  get,
  path = "/v0/characters/{character_id}",
  params(("character_id" = u32, Path, description = "角色 ID")),
  responses(
    (status = 200, description = "角色详情", body = CharacterDoc),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_character(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path(character_id): Path<u32>,
) -> ApiResult<CharacterDoc> {
  let allow_nsfw = auth.allow_nsfw;

  let mut qb = QueryBuilder::new(
    "SELECT c.crt_id, c.crt_name, c.crt_role, c.crt_summary, c.crt_img, c.crt_comment, c.crt_collects, c.crt_lock, \
            f.gender, f.bloodtype, f.birth_year, f.birth_mon, f.birth_day \
     FROM chii_characters c \
     LEFT JOIN chii_person_fields f ON f.prsn_cat = 'crt' AND f.prsn_id = c.crt_id \
     WHERE c.crt_redirect = 0 AND c.crt_id = ",
  );
  qb.push_bind(character_id);
  if !allow_nsfw {
    qb.push(" AND c.crt_nsfw = 0");
  }
  qb.push(" LIMIT 1");

  let row: Option<CharacterRow> = qb
    .build_query_as()
    .fetch_optional(&state.pool)
    .await
    .map_err(|_| super::ApiError::internal("load character failed"))?;

  let row = row.ok_or_else(|| super::ApiError::not_found("character not found"))?;
  Ok(Json(character_from_row(&row)))
}

#[utoipa::path(
  get,
  path = "/v0/characters/{character_id}/image",
  params(
    ("character_id" = u32, Path, description = "角色 ID"),
    ("type" = String, Query, description = "图片尺寸，可选值：small, grid, large, medium")
  ),
  responses(
    (status = 302, description = "重定向到图片地址"),
    (status = 400, description = "请求参数错误", body = super::ErrorBody),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_character_image(
  State(state): State<AppState>,
  Path(character_id): Path<u32>,
  Query(query): Query<ImageQuery>,
) -> Result<Redirect, super::ApiError> {
  let repo = DbCharacterImageRepo { pool: &state.pool };
  let image_url =
    resolve_character_image_url(&repo, character_id, &query.image_type).await?;
  Ok(Redirect::temporary(&image_url))
}

#[utoipa::path(
  post,
  path = "/v0/characters/{character_id}/collect",
  params(("character_id" = u32, Path, description = "角色 ID")),
  responses(
    (status = 204, description = "收藏成功"),
    (status = 401, description = "未授权", body = super::ErrorBody),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn collect_character(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path(character_id): Path<u32>,
) -> Result<StatusCode, super::ApiError> {
  let user_id = user_id_from_auth(auth)?;
  collect_character_with_pool(&state.pool, user_id, character_id).await?;
  Ok(StatusCode::NO_CONTENT)
}

#[utoipa::path(
  delete,
  path = "/v0/characters/{character_id}/collect",
  params(("character_id" = u32, Path, description = "角色 ID")),
  responses(
    (status = 204, description = "取消收藏成功"),
    (status = 401, description = "未授权", body = super::ErrorBody),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn uncollect_character(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path(character_id): Path<u32>,
) -> Result<StatusCode, super::ApiError> {
  let user_id = user_id_from_auth(auth)?;
  uncollect_character_with_pool(&state.pool, user_id, character_id).await?;
  Ok(StatusCode::NO_CONTENT)
}

#[utoipa::path(
  get,
  path = "/v0/characters/{character_id}/subjects",
  params(("character_id" = u32, Path, description = "角色 ID")),
  responses(
    (status = 200, description = "角色关联条目", body = Vec<CharacterRelatedSubject>),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_character_related_subjects(
  State(state): State<AppState>,
  Path(character_id): Path<u32>,
) -> ApiResult<Vec<CharacterRelatedSubject>> {
  ensure_character_exists(&state.pool, character_id).await?;

  let rows: Vec<RelatedSubjectRow> = sqlx::query_as(
    "SELECT i.subject_id, s.subject_type_id AS subject_type, s.subject_name, s.subject_name_cn, \
            s.subject_image, i.crt_type AS relation_type, i.crt_appear_eps AS eps \
     FROM chii_crt_subject_index i \
     JOIN chii_subjects s ON s.subject_id = i.subject_id \
     JOIN chii_subject_fields f ON f.field_sid = s.subject_id \
     WHERE i.crt_id = ? AND s.subject_ban = 0 AND f.field_redirect = 0 \
     ORDER BY i.crt_order, s.subject_id",
  )
  .bind(character_id)
  .fetch_all(&state.pool)
  .await
  .map_err(|_| super::ApiError::internal("load related subjects failed"))?;

  let data = rows
    .into_iter()
    .map(|row| CharacterRelatedSubject {
      id: row.subject_id,
      subject_type: row.subject_type,
      staff: character_staff_string(row.relation_type),
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
  path = "/v0/characters/{character_id}/persons",
  params(("character_id" = u32, Path, description = "角色 ID")),
  responses(
    (status = 200, description = "角色关联人物", body = Vec<CharacterRelatedPerson>),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_character_related_persons(
  State(state): State<AppState>,
  Path(character_id): Path<u32>,
) -> ApiResult<Vec<CharacterRelatedPerson>> {
  ensure_character_exists(&state.pool, character_id).await?;

  let rows: Vec<RelatedPersonRow> = sqlx::query_as(
    "SELECT p.prsn_id AS person_id, p.prsn_name AS person_name, p.prsn_type AS person_type, p.prsn_img AS person_img, \
            s.subject_id, s.subject_type_id AS subject_type, s.subject_name, s.subject_name_cn, \
            COALESCE(si.crt_type, 0) AS relation_type \
     FROM chii_crt_cast_index ci \
     JOIN chii_persons p ON p.prsn_id = ci.prsn_id \
     JOIN chii_subjects s ON s.subject_id = ci.subject_id \
     JOIN chii_subject_fields f ON f.field_sid = s.subject_id \
     LEFT JOIN chii_crt_subject_index si ON si.crt_id = ci.crt_id AND si.subject_id = ci.subject_id \
     WHERE ci.crt_id = ? AND p.prsn_redirect = 0 AND s.subject_ban = 0 AND f.field_redirect = 0 \
     ORDER BY s.subject_id, p.prsn_id",
  )
  .bind(character_id)
  .fetch_all(&state.pool)
  .await
  .map_err(|_| super::ApiError::internal("load related persons failed"))?;

  let data = rows
    .into_iter()
    .map(|row| CharacterRelatedPerson {
      id: row.person_id,
      name: row.person_name,
      person_type: row.person_type,
      images: person_image(&row.person_img),
      subject_id: row.subject_id,
      subject_type: row.subject_type,
      subject_name: row.subject_name,
      subject_name_cn: row.subject_name_cn,
      staff: character_staff_string(row.relation_type),
    })
    .collect();

  Ok(Json(data))
}

async fn ensure_character_exists<'e>(
  executor: impl MySqlExecutor<'e>,
  character_id: u32,
) -> Result<(), super::ApiError> {
  let exists = sqlx::query_scalar::<_, u32>(
    "SELECT crt_id FROM chii_characters WHERE crt_redirect = 0 AND crt_id = ? LIMIT 1",
  )
  .bind(character_id)
  .fetch_optional(executor)
  .await
  .map_err(|_| super::ApiError::internal("load character failed"))?
  .is_some();

  if !exists {
    return Err(super::ApiError::not_found("character not found"));
  }

  Ok(())
}

pub(super) async fn collect_character_with_pool(
  pool: &sqlx::MySqlPool,
  user_id: u32,
  character_id: u32,
) -> Result<(), super::ApiError> {
  ensure_character_exists(pool, character_id).await?;

  let exists = sqlx::query_scalar::<_, u32>(
    "SELECT prsn_clt_id FROM chii_person_collects WHERE prsn_clt_cat = 'crt' AND prsn_clt_uid = ? AND prsn_clt_mid = ? LIMIT 1",
  )
  .bind(user_id)
  .bind(character_id)
  .fetch_optional(pool)
  .await
  .map_err(|_| super::ApiError::internal("query character collect failed"))?
  .is_some();

  if exists {
    return Ok(());
  }

  sqlx::query(
    "INSERT INTO chii_person_collects (prsn_clt_cat, prsn_clt_mid, prsn_clt_uid, prsn_clt_dateline) VALUES ('crt', ?, ?, UNIX_TIMESTAMP())",
  )
  .bind(character_id)
  .bind(user_id)
  .execute(pool)
  .await
  .map_err(|_| super::ApiError::internal("add character collect failed"))?;

  Ok(())
}

pub(super) async fn uncollect_character_with_pool(
  pool: &sqlx::MySqlPool,
  user_id: u32,
  character_id: u32,
) -> Result<(), super::ApiError> {
  ensure_character_exists(pool, character_id).await?;

  let result = sqlx::query(
    "DELETE FROM chii_person_collects WHERE prsn_clt_cat = 'crt' AND prsn_clt_uid = ? AND prsn_clt_mid = ?",
  )
  .bind(user_id)
  .bind(character_id)
  .execute(pool)
  .await
  .map_err(|_| super::ApiError::internal("remove character collect failed"))?;

  if result.rows_affected() == 0 {
    return Err(super::ApiError::not_found("character not collected"));
  }

  Ok(())
}

#[cfg(test)]
#[allow(dead_code)]
pub(super) async fn collect_character_with_tx(
  tx: &mut sqlx::Transaction<'_, sqlx::MySql>,
  user_id: u32,
  character_id: u32,
) -> Result<(), super::ApiError> {
  ensure_character_exists(&mut **tx, character_id).await?;

  let exists = sqlx::query_scalar::<_, u32>(
    "SELECT prsn_clt_id FROM chii_person_collects WHERE prsn_clt_cat = 'crt' AND prsn_clt_uid = ? AND prsn_clt_mid = ? LIMIT 1",
  )
  .bind(user_id)
  .bind(character_id)
  .fetch_optional(&mut **tx)
  .await
  .map_err(|_| super::ApiError::internal("query character collect failed"))?
  .is_some();

  if exists {
    return Ok(());
  }

  sqlx::query(
    "INSERT INTO chii_person_collects (prsn_clt_cat, prsn_clt_mid, prsn_clt_uid, prsn_clt_dateline) VALUES ('crt', ?, ?, UNIX_TIMESTAMP())",
  )
  .bind(character_id)
  .bind(user_id)
  .execute(&mut **tx)
  .await
  .map_err(|_| super::ApiError::internal("add character collect failed"))?;

  Ok(())
}

#[cfg(test)]
#[allow(dead_code)]
pub(super) async fn uncollect_character_with_tx(
  tx: &mut sqlx::Transaction<'_, sqlx::MySql>,
  user_id: u32,
  character_id: u32,
) -> Result<(), super::ApiError> {
  ensure_character_exists(&mut **tx, character_id).await?;

  let result = sqlx::query(
    "DELETE FROM chii_person_collects WHERE prsn_clt_cat = 'crt' AND prsn_clt_uid = ? AND prsn_clt_mid = ?",
  )
  .bind(user_id)
  .bind(character_id)
  .execute(&mut **tx)
  .await
  .map_err(|_| super::ApiError::internal("remove character collect failed"))?;

  if result.rows_affected() == 0 {
    return Err(super::ApiError::not_found("character not collected"));
  }

  Ok(())
}

async fn resolve_character_image_url(
  repo: &impl CharacterImageRepo,
  character_id: u32,
  image_type: &str,
) -> Result<String, super::ApiError> {
  let path = repo.find_character_image_path(character_id).await?;
  let path = path.ok_or_else(|| super::ApiError::not_found("character not found"))?;

  let image_url = select_person_image_url(&path, image_type).ok_or_else(|| {
    super::ApiError::bad_request(format!("bad image type: {image_type}"))
  })?;

  if image_url.is_empty() {
    return Ok(DEFAULT_IMAGE_URL.to_string());
  }

  Ok(image_url)
}

async fn load_characters(
  state: &AppState,
  ids: &[u32],
) -> Result<Vec<CharacterDoc>, super::ApiError> {
  if ids.is_empty() {
    return Ok(Vec::new());
  }

  let mut qb = QueryBuilder::new(
    "SELECT c.crt_id, c.crt_name, c.crt_role, c.crt_summary, c.crt_img, c.crt_comment, c.crt_collects, c.crt_lock, \
            f.gender, f.bloodtype, f.birth_year, f.birth_mon, f.birth_day \
     FROM chii_characters c \
     LEFT JOIN chii_person_fields f ON f.prsn_cat = 'crt' AND f.prsn_id = c.crt_id \
     WHERE c.crt_redirect = 0 AND c.crt_id IN (",
  );

  {
    let mut separated = qb.separated(", ");
    for id in ids {
      separated.push_bind(*id);
    }
  }
  qb.push(")");

  let rows: Vec<CharacterRow> = qb
    .build_query_as()
    .fetch_all(&state.pool)
    .await
    .map_err(|_| super::ApiError::internal("load characters failed"))?;

  let by_id: HashMap<u32, CharacterRow> =
    rows.into_iter().map(|x| (x.crt_id, x)).collect();

  let mut result = Vec::with_capacity(ids.len());
  for id in ids {
    if let Some(row) = by_id.get(id) {
      result.push(character_from_row(row));
    }
  }

  Ok(result)
}

fn character_from_row(row: &CharacterRow) -> CharacterDoc {
  CharacterDoc {
    id: row.crt_id,
    name: row.crt_name.clone(),
    role: row.crt_role,
    summary: row.crt_summary.clone(),
    locked: row.crt_lock != 0,
    images: person_image(&row.crt_img),
    stat: Stat {
      comments: row.crt_comment,
      collects: row.crt_collects,
    },
    gender: map_gender(row.gender),
    blood_type: row.bloodtype.filter(|x| *x != 0),
    birth_year: row.birth_year.filter(|x| *x != 0),
    birth_mon: row.birth_mon.filter(|x| *x != 0),
    birth_day: row.birth_day.filter(|x| *x != 0),
  }
}

fn map_gender(raw: Option<u8>) -> Option<String> {
  match raw {
    Some(1) => Some("male".to_string()),
    Some(2) => Some("female".to_string()),
    _ => None,
  }
}

#[cfg(test)]
mod tests {
  use super::resolve_character_image_url;
  use crate::server::test_mocks::MockPool;
  use axum::http::StatusCode;

  #[tokio::test]
  async fn resolve_character_image_url_returns_default_for_empty_path() {
    let mut pool = MockPool::new();
    pool
      .character_image_repo
      .expect_find_character_image_path()
      .withf(|character_id| *character_id == 8)
      .times(1)
      .returning(|_| Ok(Some(String::new())));

    let got = resolve_character_image_url(&pool.character_image_repo, 8, "small")
      .await
      .expect("resolve image");

    assert_eq!(got, "https://lain.bgm.tv/img/no_icon_subject.png");
  }

  #[tokio::test]
  async fn resolve_character_image_url_returns_not_found_when_missing() {
    let mut pool = MockPool::new();
    pool
      .character_image_repo
      .expect_find_character_image_path()
      .withf(|character_id| *character_id == 404)
      .times(1)
      .returning(|_| Ok(None));

    let err = resolve_character_image_url(&pool.character_image_repo, 404, "small")
      .await
      .expect_err("expect not found");

    assert_eq!(err.status, StatusCode::NOT_FOUND);
    assert_eq!(err.message, "character not found");
  }

  #[tokio::test]
  async fn resolve_character_image_url_returns_bad_request_for_invalid_type() {
    let mut pool = MockPool::new();
    pool
      .character_image_repo
      .expect_find_character_image_path()
      .times(1)
      .returning(|_| Ok(Some("ab/cd.jpg".to_string())));

    let err = resolve_character_image_url(&pool.character_image_repo, 1, "invalid")
      .await
      .expect_err("expect bad request");

    assert_eq!(err.status, StatusCode::BAD_REQUEST);
    assert_eq!(err.message, "bad image type: invalid");
  }
}
