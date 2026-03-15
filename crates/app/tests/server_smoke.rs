mod common;

use axum::body::Body;
use axum::http::{Request, StatusCode};
use tower::util::ServiceExt;

#[tokio::test]
async fn openapi_json_is_available() {
  let app = common::test_router().await.expect("build test router");

  let response = app
    .oneshot(
      Request::builder()
        .uri("/openapi.json")
        .method("GET")
        .body(Body::empty())
        .expect("build request"),
    )
    .await
    .expect("perform request");

  assert_eq!(response.status(), StatusCode::OK);
}

#[tokio::test]
async fn image_routes_require_type_query_like_go() {
  let app = common::test_router().await.expect("build test router");

  let cases = [
    "/v0/subjects/1/image",
    "/v0/characters/1/image",
    "/v0/persons/1/image",
  ];

  for uri in cases {
    let response = app
      .clone()
      .oneshot(
        Request::builder()
          .uri(uri)
          .method("GET")
          .body(Body::empty())
          .expect("build request"),
      )
      .await
      .expect("perform request");

    assert_eq!(response.status(), StatusCode::BAD_REQUEST, "uri: {uri}");
  }
}

#[tokio::test]
async fn image_routes_reject_non_numeric_id_like_go() {
  let app = common::test_router().await.expect("build test router");

  let cases = [
    "/v0/subjects/abc/image?type=small",
    "/v0/characters/abc/image?type=small",
    "/v0/persons/abc/image?type=small",
  ];

  for uri in cases {
    let response = app
      .clone()
      .oneshot(
        Request::builder()
          .uri(uri)
          .method("GET")
          .body(Body::empty())
          .expect("build request"),
      )
      .await
      .expect("perform request");

    assert_eq!(response.status(), StatusCode::BAD_REQUEST, "uri: {uri}");
  }
}

#[tokio::test]
async fn subject_related_routes_reject_non_numeric_id_like_go() {
  let app = common::test_router().await.expect("build test router");

  let cases = [
    "/v0/subjects/abc/persons",
    "/v0/subjects/abc/characters",
    "/v0/subjects/abc/subjects",
  ];

  for uri in cases {
    let response = app
      .clone()
      .oneshot(
        Request::builder()
          .uri(uri)
          .method("GET")
          .body(Body::empty())
          .expect("build request"),
      )
      .await
      .expect("perform request");

    assert_eq!(response.status(), StatusCode::BAD_REQUEST, "uri: {uri}");
  }
}

#[tokio::test]
async fn character_person_related_routes_reject_non_numeric_id_like_go() {
  let app = common::test_router().await.expect("build test router");

  let cases = [
    "/v0/characters/abc/subjects",
    "/v0/characters/abc/persons",
    "/v0/persons/abc/subjects",
    "/v0/persons/abc/characters",
  ];

  for uri in cases {
    let response = app
      .clone()
      .oneshot(
        Request::builder()
          .uri(uri)
          .method("GET")
          .body(Body::empty())
          .expect("build request"),
      )
      .await
      .expect("perform request");

    assert_eq!(response.status(), StatusCode::BAD_REQUEST, "uri: {uri}");
  }
}

#[tokio::test]
async fn character_collect_requires_auth_like_go() {
  let app = common::test_router().await.expect("build test router");

  let response = app
    .clone()
    .oneshot(
      Request::builder()
        .uri("/v0/characters/1/collect")
        .method("POST")
        .body(Body::empty())
        .expect("build request"),
    )
    .await
    .expect("perform request");

  assert_eq!(response.status(), StatusCode::UNAUTHORIZED);
}

#[tokio::test]
async fn character_collect_rejects_non_numeric_id_like_go() {
  let app = common::test_router().await.expect("build test router");

  let response = app
    .clone()
    .oneshot(
      Request::builder()
        .uri("/v0/characters/abc/collect")
        .method("POST")
        .body(Body::empty())
        .expect("build request"),
    )
    .await
    .expect("perform request");

  assert_eq!(response.status(), StatusCode::BAD_REQUEST);
}

#[tokio::test]
async fn person_collect_requires_auth_like_go() {
  let app = common::test_router().await.expect("build test router");

  let response = app
    .clone()
    .oneshot(
      Request::builder()
        .uri("/v0/persons/1/collect")
        .method("POST")
        .body(Body::empty())
        .expect("build request"),
    )
    .await
    .expect("perform request");

  assert_eq!(response.status(), StatusCode::UNAUTHORIZED);
}

#[tokio::test]
async fn person_collect_rejects_non_numeric_id_like_go() {
  let app = common::test_router().await.expect("build test router");

  let response = app
    .clone()
    .oneshot(
      Request::builder()
        .uri("/v0/persons/abc/collect")
        .method("POST")
        .body(Body::empty())
        .expect("build request"),
    )
    .await
    .expect("perform request");

  assert_eq!(response.status(), StatusCode::BAD_REQUEST);
}

#[tokio::test]
async fn user_collection_routes_reject_invalid_params_like_go() {
  let app = common::test_router().await.expect("build test router");

  let cases = [
    "/v0/users/test/collections?type=99",
    "/v0/users/test/collections?subject_type=99",
    "/v0/users/test/collections/abc",
  ];

  for uri in cases {
    let response = app
      .clone()
      .oneshot(
        Request::builder()
          .uri(uri)
          .method("GET")
          .body(Body::empty())
          .expect("build request"),
      )
      .await
      .expect("perform request");

    assert_eq!(response.status(), StatusCode::BAD_REQUEST, "uri: {uri}");
  }
}
