use std::net::{IpAddr, Ipv4Addr, Ipv6Addr, SocketAddr};

use axum::{
  body::Body,
  extract::{Query, State},
  http::StatusCode,
  http::{
    header::AUTHORIZATION, header::USER_AGENT, HeaderMap, HeaderName, HeaderValue,
    Request,
  },
  middleware::{self, Next},
  response::{IntoResponse, Response},
  routing::{get, post},
  Json, Router,
};
use meilisearch_sdk::{client::Client as MeiliClient, search::SearchResults};
use serde::{Deserialize, Serialize};
use sqlx::MySqlPool;
use std::{
  collections::HashMap,
  sync::OnceLock,
  time::{SystemTime, UNIX_EPOCH},
};
use tracing::Instrument;
use utoipa::OpenApi;
use uuid::Uuid;

mod characters;
mod media;
mod persons;
mod subjects;
#[cfg(test)]
mod test_mocks;
mod users;

const DEFAULT_LIMIT: usize = 10;
const MAX_LIMIT: usize = 20;
const NSFW_THRESHOLD_SECONDS: i64 = 60 * 24 * 60 * 60;

#[derive(Clone)]
pub struct AppState {
  meili: MeiliClient,
  pool: MySqlPool,
}

impl AppState {
  pub fn new(meili: MeiliClient, pool: MySqlPool) -> Self {
    Self { meili, pool }
  }

  pub fn pool(&self) -> &MySqlPool {
    &self.pool
  }
}

pub(super) trait MySqlExecutor<'e>:
  sqlx::Executor<'e, Database = sqlx::MySql>
{
}

impl<'e, T> MySqlExecutor<'e> for T where T: sqlx::Executor<'e, Database = sqlx::MySql> {}

#[derive(Debug, Clone, Copy, Default)]
pub(super) struct RequestAuth {
  pub(super) user_id: Option<u32>,
  pub(super) allow_nsfw: bool,
}

#[derive(Debug, Deserialize, utoipa::IntoParams)]
#[into_params(parameter_in = Query)]
pub(super) struct PageQuery {
  limit: Option<usize>,
  offset: Option<usize>,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, utoipa::ToSchema)]
#[serde(try_from = "u8", into = "u8")]
#[repr(u8)]
pub(super) enum SubjectType {
  Book = 1,
  Anime = 2,
  Music = 3,
  Game = 4,
  Real = 6,
}

impl TryFrom<u8> for SubjectType {
  type Error = &'static str;

  fn try_from(value: u8) -> Result<Self, Self::Error> {
    match value {
      1 => Ok(Self::Book),
      2 => Ok(Self::Anime),
      3 => Ok(Self::Music),
      4 => Ok(Self::Game),
      6 => Ok(Self::Real),
      _ => Err("invalid subject type"),
    }
  }
}

impl From<SubjectType> for u8 {
  fn from(value: SubjectType) -> Self {
    value as u8
  }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, utoipa::ToSchema)]
#[serde(try_from = "u8", into = "u8")]
#[repr(u8)]
pub(super) enum SubjectCollectionType {
  Wish = 1,
  Done = 2,
  Doing = 3,
  OnHold = 4,
  Dropped = 5,
}

impl TryFrom<u8> for SubjectCollectionType {
  type Error = &'static str;

  fn try_from(value: u8) -> Result<Self, Self::Error> {
    match value {
      1 => Ok(Self::Wish),
      2 => Ok(Self::Done),
      3 => Ok(Self::Doing),
      4 => Ok(Self::OnHold),
      5 => Ok(Self::Dropped),
      _ => Err("invalid collection type"),
    }
  }
}

impl From<SubjectCollectionType> for u8 {
  fn from(value: SubjectCollectionType) -> Self {
    value as u8
  }
}

#[derive(Debug, Serialize, utoipa::ToSchema)]
pub(super) struct ErrorBody {
  error: String,
}

#[derive(Debug, Serialize, utoipa::ToSchema)]
pub(super) struct PageInfo {
  total: usize,
  limit: usize,
  offset: usize,
}

impl PageInfo {
  pub(super) fn new(total: usize, limit: usize, offset: usize) -> Self {
    Self {
      total,
      limit,
      offset,
    }
  }
}

#[derive(OpenApi)]
#[openapi(paths(
  subjects::search_subjects,
  subjects::get_subject,
  subjects::get_subject_image,
  subjects::get_subject_related_persons,
  subjects::get_subject_related_characters,
  subjects::get_subject_related_subjects,
  characters::search_characters,
  characters::get_character,
  characters::get_character_image,
  characters::collect_character,
  characters::uncollect_character,
  characters::get_character_related_subjects,
  characters::get_character_related_persons,
  persons::search_persons,
  persons::get_person,
  persons::get_person_image,
  persons::collect_person,
  persons::uncollect_person,
  persons::get_person_related_subjects,
  persons::get_person_related_characters,
  users::list_user_collections,
  users::get_user_collection,
))]
struct ApiDoc;

#[derive(Debug)]
pub(super) struct ApiError {
  status: StatusCode,
  message: String,
}

impl ApiError {
  pub(super) fn bad_request(message: impl Into<String>) -> Self {
    Self {
      status: StatusCode::BAD_REQUEST,
      message: message.into(),
    }
  }

  pub(super) fn internal(message: impl Into<String>) -> Self {
    Self {
      status: StatusCode::INTERNAL_SERVER_ERROR,
      message: message.into(),
    }
  }

  pub(super) fn not_found(message: impl Into<String>) -> Self {
    Self {
      status: StatusCode::NOT_FOUND,
      message: message.into(),
    }
  }

  pub(super) fn unauthorized(message: impl Into<String>) -> Self {
    Self {
      status: StatusCode::UNAUTHORIZED,
      message: message.into(),
    }
  }
}

impl IntoResponse for ApiError {
  fn into_response(self) -> Response {
    (
      self.status,
      Json(ErrorBody {
        error: self.message,
      }),
    )
      .into_response()
  }
}

pub(super) type ApiResult<T> = Result<Json<T>, ApiError>;

pub async fn state_from_env() -> anyhow::Result<AppState> {
  let meili_url = std::env::var("RUST_MEILISEARCH_URL")
    .or_else(|_| std::env::var("MEILISEARCH_URL"))
    .unwrap_or_default();
  if meili_url.trim().is_empty() {
    anyhow::bail!("missing env RUST_MEILISEARCH_URL or MEILISEARCH_URL")
  }

  let meili_key = std::env::var("RUST_MEILISEARCH_KEY")
    .or_else(|_| std::env::var("MEILISEARCH_KEY"))
    .ok();

  let mysql_dsn = std::env::var("RUST_MYSQL_DSN")
    .or_else(|_| std::env::var("MYSQL_DSN"))
    .map_err(|_| anyhow::anyhow!("missing env RUST_MYSQL_DSN or MYSQL_DSN"))?;
  let pool = sqlx::mysql::MySqlPoolOptions::new()
    .max_connections(5)
    .connect(&mysql_dsn)
    .await
    .map_err(|e| anyhow::anyhow!(e))?;

  let meili = MeiliClient::new(meili_url.trim_end_matches('/'), meili_key)
    .map_err(|e| anyhow::anyhow!(e))?;

  Ok(AppState::new(meili, pool))
}

pub fn build_router(state: AppState) -> Router {
  Router::new()
    .route("/openapi.json", get(openapi_json))
    .route("/v0/search/subjects", post(subjects::search_subjects))
    .route("/v0/search/characters", post(characters::search_characters))
    .route("/v0/search/persons", post(persons::search_persons))
    .route("/v0/subjects/:subject_id", get(subjects::get_subject))
    .route(
      "/v0/subjects/:subject_id/image",
      get(subjects::get_subject_image),
    )
    .route(
      "/v0/subjects/:subject_id/persons",
      get(subjects::get_subject_related_persons),
    )
    .route(
      "/v0/subjects/:subject_id/characters",
      get(subjects::get_subject_related_characters),
    )
    .route(
      "/v0/subjects/:subject_id/subjects",
      get(subjects::get_subject_related_subjects),
    )
    .route(
      "/v0/characters/:character_id",
      get(characters::get_character),
    )
    .route(
      "/v0/characters/:character_id/image",
      get(characters::get_character_image),
    )
    .route(
      "/v0/characters/:character_id/collect",
      post(characters::collect_character).delete(characters::uncollect_character),
    )
    .route(
      "/v0/characters/:character_id/subjects",
      get(characters::get_character_related_subjects),
    )
    .route(
      "/v0/characters/:character_id/persons",
      get(characters::get_character_related_persons),
    )
    .route("/v0/persons/:person_id", get(persons::get_person))
    .route(
      "/v0/persons/:person_id/image",
      get(persons::get_person_image),
    )
    .route(
      "/v0/persons/:person_id/collect",
      post(persons::collect_person).delete(persons::uncollect_person),
    )
    .route(
      "/v0/persons/:person_id/subjects",
      get(persons::get_person_related_subjects),
    )
    .route(
      "/v0/persons/:person_id/characters",
      get(persons::get_person_related_characters),
    )
    .route(
      "/v0/users/:username/collections",
      get(users::list_user_collections),
    )
    .route(
      "/v0/users/:username/collections/:subject_id",
      get(users::get_user_collection),
    )
    .layer(middleware::from_fn_with_state(
      state.clone(),
      request_log_context_middleware,
    ))
    .with_state(state)
}

pub async fn run() -> anyhow::Result<()> {
  let bind =
    std::env::var("RUST_HTTP_ADDR").unwrap_or_else(|_| "127.0.0.1:3000".to_string());
  let addr: SocketAddr = bind.parse()?;
  let state = state_from_env().await?;
  let app = build_router(state);

  let listener = tokio::net::TcpListener::bind(addr).await?;
  tracing::info!(%addr, "rust search server listening");

  axum::serve(
    listener,
    app.into_make_service_with_connect_info::<SocketAddr>(),
  )
  .await?;
  Ok(())
}

async fn openapi_json() -> Json<utoipa::openapi::OpenApi> {
  Json(ApiDoc::openapi())
}

async fn request_log_context_middleware(
  State(state): State<AppState>,
  mut req: Request<Body>,
  next: Next,
) -> Response {
  let auth = request_auth_from_headers(&state, req.headers()).await;
  req.extensions_mut().insert(auth);

  let request_id = request_id_from_header(req.headers().get("Cf-Ray"));
  let user_agent = req
    .headers()
    .get(USER_AGENT)
    .and_then(|h| h.to_str().ok())
    .unwrap_or("")
    .to_owned();
  let remote_addr = extract_client_ip(&req)
    .map(|x| x.to_string())
    .unwrap_or_default();

  let method = req.method().clone();
  let path = req.uri().path().to_owned();

  let span = tracing::info_span!(
    "http.request",
    request_id = %request_id,
    user_id = ?auth.user_id,
    allow_nsfw = auth.allow_nsfw,
    user_agent = %user_agent,
    remote_addr = %remote_addr,
    method = %method,
    path = %path,
  );

  let mut response = next.run(req).instrument(span).await;

  if let Ok(v) = HeaderValue::from_str(&request_id) {
    response
      .headers_mut()
      .insert(HeaderName::from_static("cf-ray"), v.clone());
    response
      .headers_mut()
      .insert(HeaderName::from_static("x-request-id"), v);
  }

  response
}

#[derive(Debug, Clone)]
enum TrustedNet {
  V4(Ipv4Addr, u8),
  V6(Ipv6Addr, u8),
}

static TRUSTED_PROXIES: OnceLock<Vec<TrustedNet>> = OnceLock::new();

fn trusted_proxies() -> &'static [TrustedNet] {
  TRUSTED_PROXIES
    .get_or_init(|| {
      std::env::var("RUST_TRUSTED_PROXIES")
        .ok()
        .into_iter()
        .flat_map(|raw| {
          raw
            .split(',')
            .map(str::trim)
            .map(ToOwned::to_owned)
            .collect::<Vec<_>>()
        })
        .filter_map(|s| parse_trusted_net(&s))
        .collect()
    })
    .as_slice()
}

fn parse_trusted_net(raw: &str) -> Option<TrustedNet> {
  if raw.is_empty() {
    return None;
  }

  let (ip_str, prefix) = if let Some((ip, p)) = raw.split_once('/') {
    let prefix = p.parse::<u8>().ok()?;
    (ip, Some(prefix))
  } else {
    (raw, None)
  };

  let ip = ip_str.parse::<IpAddr>().ok()?;
  match ip {
    IpAddr::V4(v4) => Some(TrustedNet::V4(v4, prefix.unwrap_or(32))),
    IpAddr::V6(v6) => Some(TrustedNet::V6(v6, prefix.unwrap_or(128))),
  }
}

fn extract_client_ip(req: &Request<Body>) -> Option<IpAddr> {
  let peer_ip = req
    .extensions()
    .get::<axum::extract::ConnectInfo<SocketAddr>>()
    .map(|x| x.0.ip())?;

  if !is_trusted_proxy(peer_ip) {
    return Some(peer_ip);
  }

  if let Some(ip) = header_ip(req, "cf-connecting-ip") {
    return Some(ip);
  }

  if let Some(ip) = forwarded_for_client_ip(req, peer_ip) {
    return Some(ip);
  }

  if let Some(ip) = header_ip(req, "x-real-ip") {
    return Some(ip);
  }

  Some(peer_ip)
}

fn header_ip(req: &Request<Body>, name: &str) -> Option<IpAddr> {
  req
    .headers()
    .get(name)
    .and_then(|v| v.to_str().ok())
    .and_then(parse_header_ip)
}

fn parse_header_ip(raw: &str) -> Option<IpAddr> {
  let s = raw.trim();
  if s.is_empty() {
    return None;
  }

  if let Ok(ip) = s.parse::<IpAddr>() {
    return Some(ip);
  }

  if s.starts_with('[') && s.ends_with(']') {
    return s[1..s.len() - 1].parse::<IpAddr>().ok();
  }

  None
}

fn forwarded_for_client_ip(req: &Request<Body>, peer_ip: IpAddr) -> Option<IpAddr> {
  let header = req.headers().get("x-forwarded-for")?.to_str().ok()?;
  let mut chain: Vec<IpAddr> = header.split(',').filter_map(parse_header_ip).collect();

  chain.push(peer_ip);

  chain.into_iter().rev().find(|&ip| !is_trusted_proxy(ip))
}

fn is_trusted_proxy(ip: IpAddr) -> bool {
  trusted_proxies().iter().any(|net| ip_in_net(ip, net))
}

fn ip_in_net(ip: IpAddr, net: &TrustedNet) -> bool {
  match (ip, net) {
    (IpAddr::V4(ip), TrustedNet::V4(base, prefix)) => {
      let p = (*prefix).min(32);
      let ip_u = u32::from(ip);
      let base_u = u32::from(*base);
      let mask = if p == 0 { 0 } else { u32::MAX << (32 - p) };
      (ip_u & mask) == (base_u & mask)
    }
    (IpAddr::V6(ip), TrustedNet::V6(base, prefix)) => {
      let p = (*prefix).min(128);
      let ip_u = u128::from_be_bytes(ip.octets());
      let base_u = u128::from_be_bytes(base.octets());
      let mask = if p == 0 { 0 } else { u128::MAX << (128 - p) };
      (ip_u & mask) == (base_u & mask)
    }
    _ => false,
  }
}

fn request_id_from_header(value: Option<&HeaderValue>) -> String {
  value
    .and_then(|h| h.to_str().ok())
    .map(str::trim)
    .filter(|v| !v.is_empty())
    .map(ToOwned::to_owned)
    .unwrap_or_else(|| Uuid::now_v7().to_string())
}

async fn request_auth_from_headers(
  state: &AppState,
  headers: &HeaderMap,
) -> RequestAuth {
  let token = match bearer_token_from_headers(headers) {
    Ok(v) => v,
    Err(_) => return RequestAuth::default(),
  };

  let raw_user_id = match sqlx::query_scalar::<_, String>(
    r#"SELECT t.user_id
       FROM chii_oauth_access_tokens t
       WHERE t.access_token = BINARY ? AND t.expires > NOW()
       LIMIT 1"#,
  )
  .bind(token)
  .fetch_optional(&state.pool)
  .await
  {
    Ok(v) => v,
    Err(_) => return RequestAuth::default(),
  };

  let Some(raw_user_id) = raw_user_id else {
    return RequestAuth::default();
  };

  let user_id = match parse_user_id(&raw_user_id) {
    Some(v) => v,
    None => return RequestAuth::default(),
  };

  let regdate = match sqlx::query_scalar::<_, i64>(
    "SELECT regdate FROM chii_members WHERE uid = ? LIMIT 1",
  )
  .bind(user_id)
  .fetch_optional(&state.pool)
  .await
  {
    Ok(v) => v,
    Err(_) => return RequestAuth::default(),
  };

  let Some(regdate) = regdate else {
    return RequestAuth::default();
  };

  let now = match SystemTime::now().duration_since(UNIX_EPOCH) {
    Ok(v) => v.as_secs() as i64,
    Err(_) => return RequestAuth::default(),
  };

  RequestAuth {
    user_id: Some(user_id),
    allow_nsfw: now.saturating_sub(regdate) >= NSFW_THRESHOLD_SECONDS,
  }
}

fn parse_user_id(raw: &str) -> Option<u32> {
  let value = raw.trim();
  if value.is_empty() || !value.chars().all(|c| c.is_ascii_digit()) {
    return None;
  }

  value.parse::<u32>().ok()
}

fn bearer_token_from_headers(headers: &HeaderMap) -> Result<&str, ()> {
  let Some(auth_header) = headers.get(AUTHORIZATION) else {
    return Err(());
  };

  let Ok(auth_header) = auth_header.to_str() else {
    return Err(());
  };

  let Some((scheme, token)) = auth_header.split_once(' ') else {
    return Err(());
  };

  if scheme != "Bearer" || token.trim().is_empty() {
    return Err(());
  }

  Ok(token.trim())
}

pub(super) fn user_id_from_auth(auth: RequestAuth) -> Result<u32, ApiError> {
  auth
    .user_id
    .ok_or_else(|| ApiError::unauthorized("missing or invalid bearer token"))
}

#[derive(Debug, Clone, Deserialize)]
struct PlatformInfo {
  #[serde(default)]
  r#type: String,
  #[serde(default)]
  type_cn: String,
}

static PLATFORM_MAP: OnceLock<HashMap<u8, HashMap<u16, PlatformInfo>>> =
  OnceLock::new();

fn platform_map() -> &'static HashMap<u8, HashMap<u16, PlatformInfo>> {
  PLATFORM_MAP.get_or_init(|| {
    serde_json::from_str(include_str!("../../../../pkg/vars/platform.go.json"))
      .unwrap_or_default()
  })
}

#[derive(Debug, Clone, Deserialize)]
struct VarsMapRoot {
  define: VarsMapDefine,
}

#[derive(Debug, Clone, Deserialize)]
struct VarsMapDefine {
  #[serde(rename = "type")]
  type_map: HashMap<String, u8>,
  types: HashMap<String, HashMap<String, VarsMapItem>>,
}

#[derive(Debug, Clone, Deserialize)]
struct VarsMapItem {
  #[serde(default)]
  cn: String,
}

fn decode_vars_map(raw: &str) -> HashMap<u8, HashMap<u16, String>> {
  let parsed: VarsMapRoot = match serde_json::from_str(raw) {
    Ok(v) => v,
    Err(_) => return HashMap::new(),
  };

  let mut result: HashMap<u8, HashMap<u16, String>> = HashMap::new();

  for (type_name, values) in parsed.define.types {
    let Some(type_id) = parsed.define.type_map.get(&type_name).copied() else {
      continue;
    };

    let mut items = HashMap::new();
    for (relation_id, relation) in values {
      let Ok(id) = relation_id.parse::<u16>() else {
        continue;
      };
      if !relation.cn.is_empty() {
        items.insert(id, relation.cn);
      }
    }

    if !items.is_empty() {
      result.insert(type_id, items);
    }
  }

  result
}

static RELATION_MAP: OnceLock<HashMap<u8, HashMap<u16, String>>> = OnceLock::new();

fn relation_map() -> &'static HashMap<u8, HashMap<u16, String>> {
  RELATION_MAP.get_or_init(|| {
    decode_vars_map(include_str!("../../../../pkg/vars/relations.go.json"))
  })
}

static STAFF_MAP: OnceLock<HashMap<u8, HashMap<u16, String>>> = OnceLock::new();

fn staff_map() -> &'static HashMap<u8, HashMap<u16, String>> {
  STAFF_MAP.get_or_init(|| {
    decode_vars_map(include_str!("../../../../pkg/vars/staffs.go.json"))
  })
}

fn subject_type_string(subject_type: SubjectType) -> &'static str {
  match subject_type {
    SubjectType::Book => "书籍",
    SubjectType::Anime => "动画",
    SubjectType::Music => "音乐",
    SubjectType::Game => "游戏",
    SubjectType::Real => "三次元",
  }
}

fn subject_type_string_or_unknown(subject_type: u8) -> &'static str {
  SubjectType::try_from(subject_type)
    .map(subject_type_string)
    .unwrap_or("unknown subject type")
}

pub(super) fn relation_string(
  destination_subject_type: u8,
  relation_type: u16,
) -> String {
  if relation_type == 1 {
    return subject_type_string_or_unknown(destination_subject_type).to_string();
  }

  relation_map()
    .get(&destination_subject_type)
    .and_then(|m| m.get(&relation_type))
    .cloned()
    .unwrap_or_else(|| {
      subject_type_string_or_unknown(destination_subject_type).to_string()
    })
}

pub(super) fn staff_string(subject_type: u8, staff_type: u16) -> String {
  staff_map()
    .get(&subject_type)
    .and_then(|m| m.get(&staff_type))
    .cloned()
    .unwrap_or_default()
}

pub(super) fn character_staff_string(staff_type: u8) -> String {
  match staff_type {
    1 => "主角".to_string(),
    2 => "配角".to_string(),
    3 => "客串".to_string(),
    4 => "闲角".to_string(),
    5 => "旁白".to_string(),
    6 => "声库".to_string(),
    _ => String::new(),
  }
}

pub(super) fn platform_string(subject_type: u8, platform_id: u16) -> Option<String> {
  let platform = platform_map()
    .get(&subject_type)
    .and_then(|by_type| by_type.get(&platform_id));

  match platform {
    Some(v) => {
      if !v.type_cn.is_empty() {
        Some(v.type_cn.clone())
      } else if !v.r#type.is_empty() {
        Some(v.r#type.clone())
      } else {
        None
      }
    }
    None => {
      if subject_type != 0 {
        tracing::warn!(subject_type, platform_id, "unknown platform mapping");
      }
      None
    }
  }
}

pub(super) async fn execute_search<T>(
  state: &AppState,
  index: &str,
  keyword: &str,
  limit: usize,
  offset: usize,
  filter: Option<String>,
  sort: Option<&[&str]>,
) -> Result<SearchResults<T>, ApiError>
where
  T: serde::de::DeserializeOwned + Send + Sync + 'static,
{
  let index = state.meili.index(index);
  let mut query = index.search();
  query
    .with_query(keyword)
    .with_limit(limit)
    .with_offset(offset);

  if let Some(filter) = filter.as_deref() {
    query.with_filter(filter);
  }

  if let Some(sort) = sort {
    query.with_sort(sort);
  }

  query.execute().await.map_err(|e| {
    tracing::error!(error = %e, "search request failed");
    ApiError::internal("search failed")
  })
}

pub(super) fn search_total<T>(docs: &SearchResults<T>) -> usize {
  docs
    .estimated_total_hits
    .or(docs.total_hits)
    .unwrap_or(docs.hits.len())
}

pub(super) fn parse_page(page: Query<PageQuery>) -> (usize, usize) {
  let page = page.0;
  let limit = page.limit.unwrap_or(DEFAULT_LIMIT).min(MAX_LIMIT);
  let offset = page.offset.unwrap_or(0);
  (limit, offset)
}

pub(super) fn join_filter(items: &[String]) -> Option<String> {
  if items.is_empty() {
    None
  } else {
    Some(items.join(" AND "))
  }
}

pub(super) fn quote_str(value: &str) -> String {
  match serde_json::to_string(value) {
    Ok(v) => v,
    Err(_) => format!("\"{}\"", value.replace('"', "\\\"")),
  }
}

fn parse_op_and_value(input: &str) -> (&str, &str) {
  let trimmed = input.trim();
  if let Some(rest) = trimmed.strip_prefix(">=") {
    return (">=", rest.trim());
  }
  if let Some(rest) = trimmed.strip_prefix("<=") {
    return ("<=", rest.trim());
  }
  if let Some(rest) = trimmed.strip_prefix('>') {
    return (">", rest.trim());
  }
  if let Some(rest) = trimmed.strip_prefix('<') {
    return ("<", rest.trim());
  }
  if let Some(rest) = trimmed.strip_prefix('=') {
    return ("=", rest.trim());
  }
  ("=", trimmed)
}

pub(super) fn parse_integer_filter(
  input: &str,
  key: &str,
) -> Result<(String, i64), ApiError> {
  let (op, value) = parse_op_and_value(input);
  let number = value.parse::<i64>().map_err(|_| {
    ApiError::bad_request(format!(
      "invalid {key} filter: {input:?}, should be in the format of \"^(>|<|>=|<=|=) *\\d+$\""
    ))
  })?;
  Ok((op.to_string(), number))
}

pub(super) fn parse_float_filter(
  input: &str,
  key: &str,
) -> Result<(String, f64), ApiError> {
  let (op, value) = parse_op_and_value(input);
  let number = value.parse::<f64>().map_err(|_| {
    ApiError::bad_request(format!(
      "invalid {key} filter: {input:?}, should be in the format of \"^(>|<|>=|<=|=) *\\d+(\\.\\d+)?$\""
    ))
  })?;
  Ok((op.to_string(), number))
}

pub(super) fn parse_date_filter(input: &str) -> Result<(String, i32), ApiError> {
  let (op, value) = parse_op_and_value(input);
  let value = parse_ymd_to_int(value).ok_or_else(|| {
    ApiError::bad_request(format!(
      "invalid date filter: {input:?}, date should be in the format of \"YYYY-MM-DD\""
    ))
  })?;
  Ok((op.to_string(), value))
}

fn parse_ymd_to_int(date: &str) -> Option<i32> {
  if date.len() < 10 {
    return None;
  }

  let bytes = date.as_bytes();
  if bytes.get(4).copied() != Some(b'-') || bytes.get(7).copied() != Some(b'-') {
    return None;
  }

  if !date[0..4].chars().all(|c| c.is_ascii_digit())
    || !date[5..7].chars().all(|c| c.is_ascii_digit())
    || !date[8..10].chars().all(|c| c.is_ascii_digit())
  {
    return None;
  }

  let year = date[0..4].parse::<i32>().ok()?;
  let month = date[5..7].parse::<i32>().ok()?;
  let day = date[8..10].parse::<i32>().ok()?;

  Some(year * 10000 + month * 100 + day)
}
