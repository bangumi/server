use async_trait::async_trait;
use axum::{
  extract::{Extension, Path, Query, State},
  response::Redirect,
  Json,
};
use php_serialize::from_str as parse_php_serialize;
use serde::{Deserialize, Serialize};
use sqlx::QueryBuilder;
use std::collections::HashMap;
use utoipa::ToSchema;

#[cfg(test)]
use mockall::automock;

use super::media::{
  person_image, select_subject_image_url, subject_image, PersonImages, SubjectImages,
  DEFAULT_IMAGE_URL,
};
use super::{
  character_staff_string, execute_search, join_filter, parse_date_filter,
  parse_float_filter, parse_integer_filter, parse_page, platform_string, quote_str,
  relation_string, search_total, staff_string, ApiError, ApiResult, AppState,
  MySqlExecutor, PageInfo, PageQuery, RequestAuth,
};

#[derive(Debug, Deserialize, Default, ToSchema)]
pub(super) struct SubjectReq {
  keyword: String,
  #[serde(default)]
  sort: String,
  #[serde(default)]
  filter: SubjectFilter,
}

#[derive(Debug, Deserialize, Default, ToSchema)]
pub(super) struct SubjectFilter {
  #[serde(default)]
  r#type: Vec<u8>,
  #[serde(default)]
  tag: Vec<String>,
  #[serde(default)]
  air_date: Vec<String>,
  #[serde(default)]
  rating: Vec<String>,
  #[serde(default)]
  rating_count: Vec<String>,
  #[serde(default)]
  rank: Vec<String>,
  #[serde(default)]
  meta_tags: Vec<String>,
  nsfw: Option<bool>,
}

#[derive(Debug, Deserialize, Serialize)]
struct SearchHit {
  id: u32,
}

#[derive(Debug, Deserialize, Serialize, ToSchema)]
pub(super) struct SubjectDoc {
  id: u32,
  #[serde(rename = "type")]
  type_id: u8,
  name: String,
  name_cn: String,
  summary: String,
  nsfw: bool,
  locked: bool,
  platform: Option<String>,
  #[serde(default)]
  meta_tags: Vec<String>,
  volumes: u32,
  eps: u32,
  series: bool,
  total_episodes: i64,
  rating: Rating,
  collection: Collection,
  #[serde(default)]
  tags: Vec<SubjectTag>,
  images: SubjectImages,
  date: Option<String>,
}

#[derive(Debug, Deserialize, Serialize, ToSchema)]
pub(super) struct SubjectTag {
  name: String,
  count: u32,
}

#[derive(Debug, Deserialize, Serialize, ToSchema)]
pub(super) struct Collection {
  wish: u32,
  collect: u32,
  doing: u32,
  on_hold: u32,
  dropped: u32,
}

#[derive(Debug, Deserialize, Serialize, ToSchema)]
pub(super) struct Rating {
  rank: u32,
  total: u32,
  score: f64,
  count: RatingCount,
}

#[derive(Debug, Deserialize, Serialize, ToSchema)]
pub(super) struct RatingCount {
  #[serde(rename = "1")]
  field1: u32,
  #[serde(rename = "2")]
  field2: u32,
  #[serde(rename = "3")]
  field3: u32,
  #[serde(rename = "4")]
  field4: u32,
  #[serde(rename = "5")]
  field5: u32,
  #[serde(rename = "6")]
  field6: u32,
  #[serde(rename = "7")]
  field7: u32,
  #[serde(rename = "8")]
  field8: u32,
  #[serde(rename = "9")]
  field9: u32,
  #[serde(rename = "10")]
  field10: u32,
}

#[derive(sqlx::FromRow)]
struct SubjectRow {
  subject_id: u32,
  subject_type_id: u8,
  subject_name: String,
  subject_name_cn: String,
  field_summary: String,
  subject_nsfw: bool,
  subject_ban: u8,
  subject_platform: u16,
  field_meta_tags: String,
  field_volumes: u32,
  field_eps: u32,
  subject_series: bool,
  subject_image: String,
  subject_wish: u32,
  subject_collect: u32,
  subject_doing: u32,
  subject_on_hold: u32,
  subject_dropped: u32,
  field_rank: u32,
  field_rate_1: u32,
  field_rate_2: u32,
  field_rate_3: u32,
  field_rate_4: u32,
  field_rate_5: u32,
  field_rate_6: u32,
  field_rate_7: u32,
  field_rate_8: u32,
  field_rate_9: u32,
  field_rate_10: u32,
  field_date: Option<String>,
  field_tags: Vec<u8>,
}

#[derive(Deserialize)]
struct SubjectTagItem {
  tag_name: Option<String>,
  tag_count: Option<u32>,
}

#[derive(Debug, Deserialize, ToSchema)]
pub(super) struct ImageQuery {
  #[serde(rename = "type")]
  image_type: String,
}

#[cfg_attr(test, automock)]
#[async_trait]
pub(super) trait SubjectImageRepo: Send + Sync {
  async fn find_subject_image_path(
    &self,
    subject_id: u32,
    allow_nsfw: bool,
  ) -> Result<Option<String>, ApiError>;
}

struct DbSubjectImageRepo<'a> {
  pool: &'a sqlx::MySqlPool,
}

#[async_trait]
impl SubjectImageRepo for DbSubjectImageRepo<'_> {
  async fn find_subject_image_path(
    &self,
    subject_id: u32,
    allow_nsfw: bool,
  ) -> Result<Option<String>, ApiError> {
    let mut qb = QueryBuilder::new(
      "SELECT s.subject_image \
       FROM chii_subjects s \
       JOIN chii_subject_fields f ON f.field_sid = s.subject_id \
       WHERE s.subject_id = ",
    );
    qb.push_bind(subject_id);
    qb.push(" AND s.subject_ban = 0 AND f.field_redirect = 0");
    if !allow_nsfw {
      qb.push(" AND s.subject_nsfw = 0");
    }
    qb.push(" LIMIT 1");

    qb.build_query_scalar()
      .fetch_optional(self.pool)
      .await
      .map_err(|_| ApiError::internal("load subject image failed"))
  }
}

#[derive(Debug, Serialize, ToSchema)]
pub(super) struct SubjectSearchResponse {
  #[serde(flatten)]
  page: PageInfo,
  data: Vec<SubjectDoc>,
}

#[derive(Debug, Serialize, ToSchema)]
pub(super) struct SubjectRelatedSubject {
  id: u32,
  #[serde(rename = "type")]
  subject_type: u8,
  name: String,
  name_cn: String,
  images: SubjectImages,
  relation: String,
}

#[derive(Debug, Serialize, ToSchema)]
pub(super) struct SubjectRelatedPerson {
  id: u32,
  name: String,
  #[serde(rename = "type")]
  person_type: u8,
  career: Vec<String>,
  images: PersonImages,
  relation: String,
  eps: String,
}

#[derive(Debug, Serialize, ToSchema)]
pub(super) struct SubjectRelatedCharacter {
  id: u32,
  name: String,
  summary: String,
  #[serde(rename = "type")]
  role: u8,
  images: PersonImages,
  relation: String,
  #[serde(default)]
  actors: Vec<SubjectActor>,
}

#[derive(Debug, Clone, Serialize, ToSchema)]
pub(super) struct SubjectActor {
  id: u32,
  name: String,
  short_summary: String,
  #[serde(rename = "type")]
  person_type: u8,
  career: Vec<String>,
  images: PersonImages,
  locked: bool,
}

#[derive(sqlx::FromRow)]
struct RelatedSubjectRow {
  relation_type: u16,
  related_subject_id: u32,
  related_subject_type_id: u8,
  related_subject_name: String,
  related_subject_name_cn: String,
  related_subject_image: String,
}

#[derive(sqlx::FromRow)]
struct RelatedPersonRow {
  person_id: u32,
  person_name: String,
  person_type: u8,
  person_img: String,
  prsn_position: u16,
  prsn_appear_eps: String,
  prsn_producer: bool,
  prsn_mangaka: bool,
  prsn_artist: bool,
  prsn_seiyu: bool,
  prsn_writer: bool,
  prsn_illustrator: bool,
  prsn_actor: bool,
}

#[derive(sqlx::FromRow)]
struct RelatedCharacterRow {
  character_id: u32,
  character_name: String,
  character_type: u8,
  character_summary: String,
  character_img: String,
  relation_type: u8,
}

#[derive(sqlx::FromRow)]
struct RelatedActorRow {
  character_id: u32,
  actor_id: u32,
  actor_name: String,
  actor_summary: String,
  actor_type: u8,
  actor_img: String,
  actor_lock: i8,
  prsn_producer: bool,
  prsn_mangaka: bool,
  prsn_artist: bool,
  prsn_seiyu: bool,
  prsn_writer: bool,
  prsn_illustrator: bool,
  prsn_actor: bool,
}

#[utoipa::path(
  post,
  path = "/v0/search/subjects",
  request_body = SubjectReq,
  params(PageQuery),
  responses(
    (status = 200, description = "返回搜索结果", body = SubjectSearchResponse),
    (status = 400, description = "请求参数错误", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]

pub(super) async fn search_subjects(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  page: Query<PageQuery>,
  Json(body): Json<SubjectReq>,
) -> ApiResult<SubjectSearchResponse> {
  let (limit, offset) = parse_page(page);
  let allow_nsfw = auth.allow_nsfw;

  let mut filters = Vec::new();

  if !body.filter.air_date.is_empty() {
    let mut or_items = Vec::new();
    for raw in &body.filter.air_date {
      let (op, value) = parse_date_filter(raw)?;
      or_items.push(format!("date {op} {value}"));
    }
    filters.push(format!("({})", or_items.join(" OR ")));
  }

  if !body.filter.r#type.is_empty() {
    let mut or_items = Vec::new();
    for t in &body.filter.r#type {
      or_items.push(format!("type = {t}"));
    }
    filters.push(format!("({})", or_items.join(" OR ")));
  }

  if !allow_nsfw || body.filter.nsfw == Some(false) {
    filters.push("(nsfw = false)".to_string());
  }

  for v in &body.filter.meta_tags {
    filters.push(format!("meta_tag = {}", quote_str(v)));
  }

  for v in &body.filter.tag {
    filters.push(format!("tag = {}", quote_str(v)));
  }

  for v in &body.filter.rank {
    let (op, num) = parse_integer_filter(v, "rank")?;
    filters.push(format!("rank {op} {num}"));
  }

  for v in &body.filter.rating {
    let (op, num) = parse_float_filter(v, "rating")?;
    filters.push(format!("score {op} {num}"));
  }

  for v in &body.filter.rating_count {
    let (op, num) = parse_integer_filter(v, "rating_count")?;
    filters.push(format!("rating_count {op} {num}"));
  }

  let sort = match body.sort.as_str() {
    "" | "match" => None,
    "score" => Some(["score:desc"].as_slice()),
    "heat" => Some(["heat:desc"].as_slice()),
    "rank" => Some(["rank:asc"].as_slice()),
    _ => return Err(ApiError::bad_request("sort not supported")),
  };

  let docs = execute_search::<SearchHit>(
    &state,
    "subjects",
    &body.keyword,
    limit,
    offset,
    join_filter(&filters),
    sort,
  )
  .await?;

  let ids: Vec<u32> = docs.hits.iter().map(|x| x.result.id).collect();
  let data = load_subjects(&state, &ids).await?;

  let total = search_total(&docs);

  Ok(Json(SubjectSearchResponse {
    page: PageInfo::new(total, limit, offset),
    data,
  }))
}

#[utoipa::path(
  get,
  path = "/v0/subjects/{subject_id}",
  params(("subject_id" = u32, Path, description = "条目 ID")),
  responses(
    (status = 200, description = "条目详情", body = SubjectDoc),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_subject(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path(subject_id): Path<u32>,
) -> ApiResult<SubjectDoc> {
  let allow_nsfw = auth.allow_nsfw;

  let mut qb = QueryBuilder::new(
    "SELECT s.subject_id, s.subject_type_id, s.subject_name, s.subject_name_cn, s.field_summary, \
            s.subject_nsfw, s.subject_ban, s.subject_platform, s.field_meta_tags, s.field_volumes, s.field_eps, \
            s.subject_series, s.subject_image, s.subject_wish, s.subject_collect, s.subject_doing, s.subject_on_hold, s.subject_dropped, \
            f.field_rank, f.field_rate_1, f.field_rate_2, f.field_rate_3, f.field_rate_4, f.field_rate_5, \
            f.field_rate_6, f.field_rate_7, f.field_rate_8, f.field_rate_9, f.field_rate_10, \
            DATE_FORMAT(f.field_date, '%Y-%m-%d') AS field_date, f.field_tags \
     FROM chii_subjects s \
     JOIN chii_subject_fields f ON f.field_sid = s.subject_id \
     WHERE s.subject_id = ",
  );
  qb.push_bind(subject_id);
  qb.push(" AND s.subject_ban = 0 AND f.field_redirect = 0");
  if !allow_nsfw {
    qb.push(" AND s.subject_nsfw = 0");
  }
  qb.push(" LIMIT 1");

  let row: Option<SubjectRow> =
    qb.build_query_as()
      .fetch_optional(&state.pool)
      .await
      .map_err(|_| ApiError::internal("load subject failed"))?;

  let row = row.ok_or_else(|| ApiError::not_found("subject not found"))?;
  Ok(Json(subject_from_row(&row)))
}

#[utoipa::path(
  get,
  path = "/v0/subjects/{subject_id}/image",
  params(
    ("subject_id" = u32, Path, description = "条目 ID"),
    ("type" = String, Query, description = "图片尺寸，可选值：small, grid, large, medium, common")
  ),
  responses(
    (status = 302, description = "重定向到图片地址"),
    (status = 400, description = "请求参数错误", body = super::ErrorBody),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_subject_image(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path(subject_id): Path<u32>,
  Query(query): Query<ImageQuery>,
) -> Result<Redirect, ApiError> {
  let allow_nsfw = auth.allow_nsfw;
  let repo = DbSubjectImageRepo { pool: &state.pool };
  let image_url =
    resolve_subject_image_url(&repo, subject_id, allow_nsfw, &query.image_type).await?;

  Ok(Redirect::temporary(&image_url))
}

#[utoipa::path(
  get,
  path = "/v0/subjects/{subject_id}/subjects",
  params(("subject_id" = u32, Path, description = "条目 ID")),
  responses(
    (status = 200, description = "关联条目", body = Vec<SubjectRelatedSubject>),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_subject_related_subjects(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path(subject_id): Path<u32>,
) -> ApiResult<Vec<SubjectRelatedSubject>> {
  let allow_nsfw = auth.allow_nsfw;
  ensure_subject_exists(&state.pool, subject_id, allow_nsfw).await?;

  let mut qb = QueryBuilder::new(
    "SELECT r.rlt_relation_type AS relation_type, \
            s.subject_id AS related_subject_id, s.subject_type_id AS related_subject_type_id, \
            s.subject_name AS related_subject_name, s.subject_name_cn AS related_subject_name_cn, s.subject_image AS related_subject_image \
     FROM chii_subject_relations r \
     JOIN chii_subjects s ON s.subject_id = r.rlt_related_subject_id \
     JOIN chii_subject_fields f ON f.field_sid = s.subject_id \
     WHERE r.rlt_subject_id = ",
  );
  qb.push_bind(subject_id);
  qb.push(" AND s.subject_ban = 0 AND f.field_redirect = 0");
  if !allow_nsfw {
    qb.push(" AND s.subject_nsfw = 0");
  }
  qb.push(" ORDER BY r.rlt_order");

  let rows: Vec<RelatedSubjectRow> =
    qb.build_query_as()
      .fetch_all(&state.pool)
      .await
      .map_err(|_| ApiError::internal("load related subjects failed"))?;

  let data = rows
    .into_iter()
    .map(|row| SubjectRelatedSubject {
      id: row.related_subject_id,
      subject_type: row.related_subject_type_id,
      name: row.related_subject_name,
      name_cn: row.related_subject_name_cn,
      images: subject_image(&row.related_subject_image),
      relation: relation_string(row.related_subject_type_id, row.relation_type),
    })
    .collect();

  Ok(Json(data))
}

#[utoipa::path(
  get,
  path = "/v0/subjects/{subject_id}/persons",
  params(("subject_id" = u32, Path, description = "条目 ID")),
  responses(
    (status = 200, description = "关联人物", body = Vec<SubjectRelatedPerson>),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_subject_related_persons(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path(subject_id): Path<u32>,
) -> ApiResult<Vec<SubjectRelatedPerson>> {
  let allow_nsfw = auth.allow_nsfw;
  let subject_type = ensure_subject_exists(&state.pool, subject_id, allow_nsfw).await?;

  let rows: Vec<RelatedPersonRow> = sqlx::query_as(
    "SELECT p.prsn_id AS person_id, p.prsn_name AS person_name, p.prsn_type AS person_type, p.prsn_img AS person_img, \
            i.prsn_position, i.prsn_appear_eps, \
            p.prsn_producer, p.prsn_mangaka, p.prsn_artist, p.prsn_seiyu, p.prsn_writer, p.prsn_illustrator, p.prsn_actor \
     FROM chii_person_cs_index i \
     JOIN chii_persons p ON p.prsn_id = i.prsn_id \
     WHERE i.subject_id = ? AND i.prsn_type = 'prsn' AND p.prsn_redirect = 0 \
     ORDER BY i.prsn_position, p.prsn_id",
  )
  .bind(subject_id)
  .fetch_all(&state.pool)
  .await
  .map_err(|_| ApiError::internal("load related persons failed"))?;

  let data = rows
    .into_iter()
    .map(|row| SubjectRelatedPerson {
      id: row.person_id,
      name: row.person_name,
      person_type: row.person_type,
      career: careers_from_flags(
        row.prsn_writer,
        row.prsn_producer,
        row.prsn_mangaka,
        row.prsn_artist,
        row.prsn_seiyu,
        row.prsn_illustrator,
        row.prsn_actor,
      ),
      images: person_image(&row.person_img),
      relation: staff_string(subject_type, row.prsn_position),
      eps: row.prsn_appear_eps,
    })
    .collect();

  Ok(Json(data))
}

#[utoipa::path(
  get,
  path = "/v0/subjects/{subject_id}/characters",
  params(("subject_id" = u32, Path, description = "条目 ID")),
  responses(
    (status = 200, description = "关联角色", body = Vec<SubjectRelatedCharacter>),
    (status = 404, description = "未找到", body = super::ErrorBody),
    (status = 500, description = "服务错误", body = super::ErrorBody)
  )
)]
pub(super) async fn get_subject_related_characters(
  State(state): State<AppState>,
  Extension(auth): Extension<RequestAuth>,
  Path(subject_id): Path<u32>,
) -> ApiResult<Vec<SubjectRelatedCharacter>> {
  let allow_nsfw = auth.allow_nsfw;
  ensure_subject_exists(&state.pool, subject_id, allow_nsfw).await?;

  let rows: Vec<RelatedCharacterRow> = sqlx::query_as(
    "SELECT c.crt_id AS character_id, c.crt_name AS character_name, c.crt_role AS character_type, \
            c.crt_summary AS character_summary, c.crt_img AS character_img, i.crt_type AS relation_type \
     FROM chii_crt_subject_index i \
     JOIN chii_characters c ON c.crt_id = i.crt_id \
     WHERE i.subject_id = ? AND c.crt_redirect = 0 \
     ORDER BY i.crt_order, c.crt_id",
  )
  .bind(subject_id)
  .fetch_all(&state.pool)
  .await
  .map_err(|_| ApiError::internal("load related characters failed"))?;

  let character_ids: Vec<u32> = rows.iter().map(|x| x.character_id).collect();
  let actors =
    load_subject_character_actors(&state, subject_id, &character_ids).await?;

  let data = rows
    .into_iter()
    .map(|row| SubjectRelatedCharacter {
      id: row.character_id,
      name: row.character_name,
      summary: row.character_summary,
      role: row.character_type,
      images: person_image(&row.character_img),
      relation: character_staff_string(row.relation_type),
      actors: actors.get(&row.character_id).cloned().unwrap_or_default(),
    })
    .collect();

  Ok(Json(data))
}

async fn resolve_subject_image_url(
  repo: &impl SubjectImageRepo,
  subject_id: u32,
  allow_nsfw: bool,
  image_type: &str,
) -> Result<String, ApiError> {
  let path = repo.find_subject_image_path(subject_id, allow_nsfw).await?;
  let path = path.ok_or_else(|| ApiError::not_found("subject not found"))?;
  let image_url = select_subject_image_url(&path, image_type)
    .ok_or_else(|| ApiError::bad_request(format!("bad image type: {image_type}")))?;

  if image_url.is_empty() {
    return Ok(DEFAULT_IMAGE_URL.to_string());
  }

  Ok(image_url)
}

async fn ensure_subject_exists<'e>(
  executor: impl MySqlExecutor<'e>,
  subject_id: u32,
  allow_nsfw: bool,
) -> Result<u8, ApiError> {
  let mut qb = QueryBuilder::new(
    "SELECT s.subject_type_id \
     FROM chii_subjects s \
     JOIN chii_subject_fields f ON f.field_sid = s.subject_id \
     WHERE s.subject_id = ",
  );
  qb.push_bind(subject_id);
  qb.push(" AND s.subject_ban = 0 AND f.field_redirect = 0");
  if !allow_nsfw {
    qb.push(" AND s.subject_nsfw = 0");
  }
  qb.push(" LIMIT 1");

  qb.build_query_scalar()
    .fetch_optional(executor)
    .await
    .map_err(|_| ApiError::internal("load subject failed"))?
    .ok_or_else(|| ApiError::not_found("subject not found"))
}

async fn load_subject_character_actors(
  state: &AppState,
  subject_id: u32,
  character_ids: &[u32],
) -> Result<HashMap<u32, Vec<SubjectActor>>, ApiError> {
  if character_ids.is_empty() {
    return Ok(HashMap::new());
  }

  let mut qb = QueryBuilder::new(
    "SELECT ci.crt_id AS character_id, \
            p.prsn_id AS actor_id, p.prsn_name AS actor_name, p.prsn_summary AS actor_summary, p.prsn_type AS actor_type, \
            p.prsn_img AS actor_img, p.prsn_lock AS actor_lock, \
            p.prsn_producer, p.prsn_mangaka, p.prsn_artist, p.prsn_seiyu, p.prsn_writer, p.prsn_illustrator, p.prsn_actor \
     FROM chii_crt_cast_index ci \
     JOIN chii_persons p ON p.prsn_id = ci.prsn_id \
     WHERE ci.subject_id = ",
  );
  qb.push_bind(subject_id);
  qb.push(" AND ci.crt_id IN (");
  {
    let mut separated = qb.separated(", ");
    for id in character_ids {
      separated.push_bind(*id);
    }
  }
  qb.push(") AND p.prsn_redirect = 0 ORDER BY ci.crt_id, p.prsn_id");

  let rows: Vec<RelatedActorRow> = qb
    .build_query_as()
    .fetch_all(&state.pool)
    .await
    .map_err(|_| ApiError::internal("load subject actors failed"))?;

  let mut by_character: HashMap<u32, Vec<SubjectActor>> = HashMap::new();
  for row in rows {
    by_character
      .entry(row.character_id)
      .or_default()
      .push(SubjectActor {
        id: row.actor_id,
        name: row.actor_name,
        short_summary: row.actor_summary,
        person_type: row.actor_type,
        career: careers_from_flags(
          row.prsn_writer,
          row.prsn_producer,
          row.prsn_mangaka,
          row.prsn_artist,
          row.prsn_seiyu,
          row.prsn_illustrator,
          row.prsn_actor,
        ),
        images: person_image(&row.actor_img),
        locked: row.actor_lock != 0,
      });
  }

  Ok(by_character)
}

fn careers_from_flags(
  writer: bool,
  producer: bool,
  mangaka: bool,
  artist: bool,
  seiyu: bool,
  illustrator: bool,
  actor: bool,
) -> Vec<String> {
  let mut items = Vec::with_capacity(7);

  if writer {
    items.push("writer".to_string());
  }
  if producer {
    items.push("producer".to_string());
  }
  if mangaka {
    items.push("mangaka".to_string());
  }
  if artist {
    items.push("artist".to_string());
  }
  if seiyu {
    items.push("seiyu".to_string());
  }
  if illustrator {
    items.push("illustrator".to_string());
  }
  if actor {
    items.push("actor".to_string());
  }

  items
}

async fn load_subjects(
  state: &AppState,
  ids: &[u32],
) -> Result<Vec<SubjectDoc>, super::ApiError> {
  if ids.is_empty() {
    return Ok(Vec::new());
  }

  let mut qb = QueryBuilder::new(
    "SELECT s.subject_id, s.subject_type_id, s.subject_name, s.subject_name_cn, s.field_summary, \
            s.subject_nsfw, s.subject_ban, s.subject_platform, s.field_meta_tags, s.field_volumes, s.field_eps, \
            s.subject_series, s.subject_image, s.subject_wish, s.subject_collect, s.subject_doing, s.subject_on_hold, s.subject_dropped, \
            f.field_rank, f.field_rate_1, f.field_rate_2, f.field_rate_3, f.field_rate_4, f.field_rate_5, \
            f.field_rate_6, f.field_rate_7, f.field_rate_8, f.field_rate_9, f.field_rate_10, \
            DATE_FORMAT(f.field_date, '%Y-%m-%d') AS field_date, f.field_tags \
     FROM chii_subjects s \
     JOIN chii_subject_fields f ON f.field_sid = s.subject_id \
     WHERE s.subject_id IN (",
  );

  {
    let mut separated = qb.separated(", ");
    for id in ids {
      separated.push_bind(*id);
    }
  }
  qb.push(") AND s.subject_ban = 0 AND f.field_redirect = 0");

  let rows: Vec<SubjectRow> = qb
    .build_query_as()
    .fetch_all(&state.pool)
    .await
    .map_err(|_| super::ApiError::internal("load subjects failed"))?;

  let by_id: HashMap<u32, SubjectRow> =
    rows.into_iter().map(|x| (x.subject_id, x)).collect();

  let mut out = Vec::with_capacity(ids.len());
  for id in ids {
    if let Some(row) = by_id.get(id) {
      out.push(subject_from_row(row));
    }
  }

  Ok(out)
}

fn subject_from_row(row: &SubjectRow) -> SubjectDoc {
  let rating = rating(row);
  SubjectDoc {
    id: row.subject_id,
    type_id: row.subject_type_id,
    name: row.subject_name.clone(),
    name_cn: row.subject_name_cn.clone(),
    summary: row.field_summary.clone(),
    nsfw: row.subject_nsfw,
    locked: row.subject_ban == 2,
    platform: platform_string(row.subject_type_id, row.subject_platform),
    meta_tags: split_meta_tags(&row.field_meta_tags),
    volumes: row.field_volumes,
    eps: row.field_eps,
    series: row.subject_series,
    total_episodes: i64::from(row.field_eps),
    rating,
    collection: Collection {
      wish: row.subject_wish,
      collect: row.subject_collect,
      doing: row.subject_doing,
      on_hold: row.subject_on_hold,
      dropped: row.subject_dropped,
    },
    tags: parse_subject_tags(&row.field_tags),
    images: subject_image(&row.subject_image),
    date: row.field_date.clone(),
  }
}

fn split_meta_tags(raw: &str) -> Vec<String> {
  raw
    .split(' ')
    .map(str::trim)
    .filter(|x| !x.is_empty())
    .map(ToOwned::to_owned)
    .collect()
}

fn parse_subject_tags(raw: &[u8]) -> Vec<SubjectTag> {
  let s = String::from_utf8_lossy(raw);
  let parsed: Vec<SubjectTagItem> = match parse_php_serialize(&s) {
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

fn rating(row: &SubjectRow) -> Rating {
  let total = row
    .field_rate_1
    .saturating_add(row.field_rate_2)
    .saturating_add(row.field_rate_3)
    .saturating_add(row.field_rate_4)
    .saturating_add(row.field_rate_5)
    .saturating_add(row.field_rate_6)
    .saturating_add(row.field_rate_7)
    .saturating_add(row.field_rate_8)
    .saturating_add(row.field_rate_9)
    .saturating_add(row.field_rate_10);

  let weighted = (row.field_rate_1 as f64) * 1.0
    + (row.field_rate_2 as f64) * 2.0
    + (row.field_rate_3 as f64) * 3.0
    + (row.field_rate_4 as f64) * 4.0
    + (row.field_rate_5 as f64) * 5.0
    + (row.field_rate_6 as f64) * 6.0
    + (row.field_rate_7 as f64) * 7.0
    + (row.field_rate_8 as f64) * 8.0
    + (row.field_rate_9 as f64) * 9.0
    + (row.field_rate_10 as f64) * 10.0;

  let score = if total == 0 {
    0.0
  } else {
    ((weighted / (total as f64)) * 10.0).round() / 10.0
  };

  Rating {
    rank: row.field_rank,
    total,
    score,
    count: RatingCount {
      field1: row.field_rate_1,
      field2: row.field_rate_2,
      field3: row.field_rate_3,
      field4: row.field_rate_4,
      field5: row.field_rate_5,
      field6: row.field_rate_6,
      field7: row.field_rate_7,
      field8: row.field_rate_8,
      field9: row.field_rate_9,
      field10: row.field_rate_10,
    },
  }
}

#[cfg(test)]
mod tests {
  use super::{resolve_subject_image_url, ApiError};
  use crate::server::test_mocks::MockPool;
  use axum::http::StatusCode;

  #[tokio::test]
  async fn resolve_subject_image_url_returns_default_for_empty_path() {
    let mut pool = MockPool::new();
    pool
      .subject_image_repo
      .expect_find_subject_image_path()
      .withf(|subject_id, allow_nsfw| *subject_id == 7 && *allow_nsfw)
      .times(1)
      .returning(|_, _| Ok(Some(String::new())));

    let got = resolve_subject_image_url(&pool.subject_image_repo, 7, true, "small")
      .await
      .expect("resolve image");

    assert_eq!(got, "https://lain.bgm.tv/img/no_icon_subject.png");
  }

  #[tokio::test]
  async fn resolve_subject_image_url_returns_not_found_when_missing() {
    let mut pool = MockPool::new();
    pool
      .subject_image_repo
      .expect_find_subject_image_path()
      .withf(|subject_id, allow_nsfw| *subject_id == 404 && !*allow_nsfw)
      .times(1)
      .returning(|_, _| Ok(None));

    let err = resolve_subject_image_url(&pool.subject_image_repo, 404, false, "small")
      .await
      .expect_err("expect not found");

    assert_eq!(err.status, StatusCode::NOT_FOUND);
    assert_eq!(err.message, "subject not found");
  }

  #[tokio::test]
  async fn resolve_subject_image_url_returns_bad_request_for_invalid_type() {
    let mut pool = MockPool::new();
    pool
      .subject_image_repo
      .expect_find_subject_image_path()
      .times(1)
      .returning(|_, _| Ok(Some("ab/cd.jpg".to_string())));

    let err = resolve_subject_image_url(&pool.subject_image_repo, 1, true, "invalid")
      .await
      .expect_err("expect bad request");

    assert_eq!(err.status, StatusCode::BAD_REQUEST);
    assert_eq!(err.message, "bad image type: invalid");
  }

  #[tokio::test]
  async fn resolve_subject_image_url_passes_through_repo_errors() {
    let mut pool = MockPool::new();
    pool
      .subject_image_repo
      .expect_find_subject_image_path()
      .times(1)
      .returning(|_, _| Err(ApiError::internal("load subject image failed")));

    let err = resolve_subject_image_url(&pool.subject_image_repo, 1, true, "small")
      .await
      .expect_err("expect internal error");

    assert_eq!(err.status, StatusCode::INTERNAL_SERVER_ERROR);
    assert_eq!(err.message, "load subject image failed");
  }
}
