mod common;

use axum::body::to_bytes;
use axum::body::Body;
use axum::http::{Request, StatusCode};
use tower::util::ServiceExt;

const DEFAULT_IMAGE_URL: &str = "https://lain.bgm.tv/img/no_icon_subject.png";

#[tokio::test]
async fn image_routes_return_404_for_nonexistent_id() {
  if !common::use_real_dependencies() {
    eprintln!("skip: RUST_TEST_REAL_DEPS is not enabled");
    return;
  }

  let app = common::test_router().await.expect("build test router");
  let missing_id = u32::MAX;

  for prefix in ["subjects", "characters", "persons"] {
    let uri = format!("/v0/{prefix}/{missing_id}/image?type=small");
    let response = app
      .clone()
      .oneshot(
        Request::builder()
          .uri(&uri)
          .method("GET")
          .body(Body::empty())
          .expect("build request"),
      )
      .await
      .expect("perform request");

    assert_eq!(response.status(), StatusCode::NOT_FOUND, "uri: {uri}");
  }
}

#[tokio::test]
async fn configured_image_fixtures_redirect() {
  if !common::use_real_dependencies() {
    eprintln!("skip: RUST_TEST_REAL_DEPS is not enabled");
    return;
  }

  let app = common::test_router().await.expect("build test router");

  assert_redirect_for_configured_fixture(
    &app,
    "subjects",
    common::env_var("RUST_TEST_SUBJECT_IMAGE_ID"),
    common::env_var("RUST_TEST_SUBJECT_IMAGE_EXPECT_DEFAULT").as_deref(),
  )
  .await;

  assert_redirect_for_configured_fixture(
    &app,
    "characters",
    common::env_var("RUST_TEST_CHARACTER_IMAGE_ID"),
    common::env_var("RUST_TEST_CHARACTER_IMAGE_EXPECT_DEFAULT").as_deref(),
  )
  .await;

  assert_redirect_for_configured_fixture(
    &app,
    "persons",
    common::env_var("RUST_TEST_PERSON_IMAGE_ID"),
    common::env_var("RUST_TEST_PERSON_IMAGE_EXPECT_DEFAULT").as_deref(),
  )
  .await;
}

#[tokio::test]
async fn related_routes_return_404_for_nonexistent_id() {
  if !common::use_real_dependencies() {
    eprintln!("skip: RUST_TEST_REAL_DEPS is not enabled");
    return;
  }

  let app = common::test_router().await.expect("build test router");
  let missing_id = u32::MAX;

  let uris = [
    format!("/v0/subjects/{missing_id}/persons"),
    format!("/v0/subjects/{missing_id}/characters"),
    format!("/v0/subjects/{missing_id}/subjects"),
    format!("/v0/characters/{missing_id}/subjects"),
    format!("/v0/characters/{missing_id}/persons"),
    format!("/v0/persons/{missing_id}/subjects"),
    format!("/v0/persons/{missing_id}/characters"),
  ];

  for uri in uris {
    let response = app
      .clone()
      .oneshot(
        Request::builder()
          .uri(&uri)
          .method("GET")
          .body(Body::empty())
          .expect("build request"),
      )
      .await
      .expect("perform request");

    assert_eq!(response.status(), StatusCode::NOT_FOUND, "uri: {uri}");
  }
}

#[tokio::test]
async fn configured_related_fixtures_return_json_array() {
  if !common::use_real_dependencies() {
    eprintln!("skip: RUST_TEST_REAL_DEPS is not enabled");
    return;
  }

  let app = common::test_router().await.expect("build test router");

  assert_related_array_for_configured_fixture(
    &app,
    "subject",
    common::env_var("RUST_TEST_SUBJECT_RELATED_ID"),
    &["persons", "characters", "subjects"],
  )
  .await;

  assert_related_array_for_configured_fixture(
    &app,
    "character",
    common::env_var("RUST_TEST_CHARACTER_RELATED_ID"),
    &["subjects", "persons"],
  )
  .await;

  assert_related_array_for_configured_fixture(
    &app,
    "person",
    common::env_var("RUST_TEST_PERSON_RELATED_ID"),
    &["subjects", "characters"],
  )
  .await;
}

#[tokio::test]
async fn write_transaction_template_can_rollback() {
  if !common::use_real_dependencies() || !common::allow_write_tests() {
    eprintln!("skip: RUST_TEST_REAL_DEPS=1 and RUST_TEST_ALLOW_WRITE=1 are required");
    return;
  }

  let state = common::test_state().await.expect("build test state");
  let mut tx = common::begin_write_transaction(&state)
    .await
    .expect("begin write transaction");

  let ping: Option<u8> = sqlx::query_scalar("SELECT 1")
    .fetch_optional(&mut *tx)
    .await
    .expect("execute in transaction");

  assert_eq!(ping, Some(1));

  tx.rollback().await.expect("rollback transaction");
}

async fn assert_redirect_for_configured_fixture(
  app: &axum::Router,
  prefix: &str,
  configured_id: Option<String>,
  expect_default: Option<&str>,
) {
  let Some(configured_id) = configured_id else {
    eprintln!("skip fixture for {prefix}: ID env var is not configured");
    return;
  };

  let id = configured_id
    .parse::<u32>()
    .expect("fixture env var must be a positive integer");
  let uri = format!("/v0/{prefix}/{id}/image?type=small");

  let response = app
    .clone()
    .oneshot(
      Request::builder()
        .uri(&uri)
        .method("GET")
        .body(Body::empty())
        .expect("build request"),
    )
    .await
    .expect("perform request");

  assert_eq!(response.status(), StatusCode::FOUND, "uri: {uri}");

  let location = response
    .headers()
    .get("location")
    .and_then(|h| h.to_str().ok())
    .unwrap_or_default();

  assert!(!location.is_empty(), "uri: {uri}, location header missing");

  let expect_default = matches!(
    expect_default,
    Some("1") | Some("true") | Some("TRUE") | Some("yes") | Some("YES")
  );
  if expect_default {
    assert_eq!(location, DEFAULT_IMAGE_URL, "uri: {uri}");
  }
}

async fn assert_related_array_for_configured_fixture(
  app: &axum::Router,
  entity: &str,
  configured_id: Option<String>,
  suffixes: &[&str],
) {
  let Some(configured_id) = configured_id else {
    eprintln!("skip fixture for {entity}: related ID env var is not configured");
    return;
  };

  let id = configured_id
    .parse::<u32>()
    .expect("fixture env var must be a positive integer");

  let base = match entity {
    "subject" => "subjects",
    "character" => "characters",
    "person" => "persons",
    _ => panic!("unsupported entity"),
  };

  for suffix in suffixes {
    let uri = format!("/v0/{base}/{id}/{suffix}");
    let response = app
      .clone()
      .oneshot(
        Request::builder()
          .uri(&uri)
          .method("GET")
          .body(Body::empty())
          .expect("build request"),
      )
      .await
      .expect("perform request");

    assert_eq!(response.status(), StatusCode::OK, "uri: {uri}");

    let body = to_bytes(response.into_body(), 1024 * 1024)
      .await
      .expect("read response body");
    let text = String::from_utf8(body.to_vec()).expect("utf8 json body");

    assert!(
      text.starts_with('[') && text.ends_with(']'),
      "uri: {uri}, body: {text}"
    );
  }
}
