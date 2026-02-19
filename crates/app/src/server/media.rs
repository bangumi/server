use serde::{Deserialize, Serialize};
use utoipa::ToSchema;

pub(super) const DEFAULT_IMAGE_URL: &str =
  "https://lain.bgm.tv/img/no_icon_subject.png";

#[derive(Debug, Clone, Default, Deserialize, Serialize, ToSchema)]
pub(super) struct PersonImages {
  small: String,
  grid: String,
  large: String,
  medium: String,
}

#[derive(Debug, Clone, Default, Deserialize, Serialize, ToSchema)]
pub(super) struct SubjectImages {
  small: String,
  grid: String,
  large: String,
  medium: String,
  common: String,
}

pub(super) fn person_image(path: &str) -> PersonImages {
  if path.is_empty() {
    return PersonImages::default();
  }

  PersonImages {
    large: format!("https://lain.bgm.tv/pic/crt/l/{path}"),
    small: format!("https://lain.bgm.tv/r/100/pic/crt/l/{path}"),
    grid: format!("https://lain.bgm.tv/pic/crt/g/{path}"),
    medium: format!("https://lain.bgm.tv/r/400/pic/crt/l/{path}"),
  }
}

pub(super) fn select_person_image_url(path: &str, image_type: &str) -> Option<String> {
  let images = person_image(path);
  match image_type {
    "small" => Some(images.small),
    "grid" => Some(images.grid),
    "large" => Some(images.large),
    "medium" => Some(images.medium),
    _ => None,
  }
}

pub(super) fn subject_image(path: &str) -> SubjectImages {
  if path.is_empty() {
    return SubjectImages::default();
  }

  SubjectImages {
    large: format!("https://lain.bgm.tv/pic/cover/l/{path}"),
    grid: format!("https://lain.bgm.tv/r/100/pic/cover/l/{path}"),
    small: format!("https://lain.bgm.tv/r/200/pic/cover/l/{path}"),
    common: format!("https://lain.bgm.tv/r/400/pic/cover/l/{path}"),
    medium: format!("https://lain.bgm.tv/r/800/pic/cover/l/{path}"),
  }
}

pub(super) fn select_subject_image_url(path: &str, image_type: &str) -> Option<String> {
  let images = subject_image(path);
  match image_type {
    "small" => Some(images.small),
    "grid" => Some(images.grid),
    "large" => Some(images.large),
    "medium" => Some(images.medium),
    "common" => Some(images.common),
    _ => None,
  }
}

#[cfg(test)]
mod tests {
  use super::{select_person_image_url, select_subject_image_url, DEFAULT_IMAGE_URL};

  #[test]
  fn select_subject_image_url_maps_all_supported_types() {
    let path = "ab/cd.jpg";

    assert_eq!(
      select_subject_image_url(path, "small").as_deref(),
      Some("https://lain.bgm.tv/r/200/pic/cover/l/ab/cd.jpg")
    );
    assert_eq!(
      select_subject_image_url(path, "grid").as_deref(),
      Some("https://lain.bgm.tv/r/100/pic/cover/l/ab/cd.jpg")
    );
    assert_eq!(
      select_subject_image_url(path, "large").as_deref(),
      Some("https://lain.bgm.tv/pic/cover/l/ab/cd.jpg")
    );
    assert_eq!(
      select_subject_image_url(path, "medium").as_deref(),
      Some("https://lain.bgm.tv/r/800/pic/cover/l/ab/cd.jpg")
    );
    assert_eq!(
      select_subject_image_url(path, "common").as_deref(),
      Some("https://lain.bgm.tv/r/400/pic/cover/l/ab/cd.jpg")
    );
  }

  #[test]
  fn select_person_image_url_maps_all_supported_types() {
    let path = "ef/gh.jpg";

    assert_eq!(
      select_person_image_url(path, "small").as_deref(),
      Some("https://lain.bgm.tv/r/100/pic/crt/l/ef/gh.jpg")
    );
    assert_eq!(
      select_person_image_url(path, "grid").as_deref(),
      Some("https://lain.bgm.tv/pic/crt/g/ef/gh.jpg")
    );
    assert_eq!(
      select_person_image_url(path, "large").as_deref(),
      Some("https://lain.bgm.tv/pic/crt/l/ef/gh.jpg")
    );
    assert_eq!(
      select_person_image_url(path, "medium").as_deref(),
      Some("https://lain.bgm.tv/r/400/pic/crt/l/ef/gh.jpg")
    );
  }

  #[test]
  fn select_image_url_rejects_invalid_type() {
    assert_eq!(select_subject_image_url("ab/cd.jpg", "x"), None);
    assert_eq!(select_person_image_url("ef/gh.jpg", "x"), None);
  }

  #[test]
  fn empty_image_path_results_in_empty_selected_url() {
    assert_eq!(select_subject_image_url("", "small").as_deref(), Some(""));
    assert_eq!(select_person_image_url("", "small").as_deref(), Some(""));
    assert_eq!(
      DEFAULT_IMAGE_URL,
      "https://lain.bgm.tv/img/no_icon_subject.png"
    );
  }
}
