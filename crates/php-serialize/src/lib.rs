use serde::de::DeserializeOwned;

pub use serde_php::Error;
pub use serde_php::Result;

pub fn from_bytes<T: DeserializeOwned>(input: &[u8]) -> Result<T> {
  serde_php::from_bytes(input)
}

pub fn from_str<T: DeserializeOwned>(input: &str) -> Result<T> {
  from_bytes(input.as_bytes())
}

#[cfg(test)]
mod tests {
  use serde::Deserialize;

  use super::from_str;

  #[derive(Debug, Deserialize, PartialEq, Eq)]
  struct TagItem {
    tag_name: Option<String>,
    result: String,
  }

  #[test]
  fn deserialize_vec_struct() {
    let raw = "a:3:{i:0;a:2:{s:8:\"tag_name\";s:6:\"动画\";s:6:\"result\";s:1:\"2\";}i:1;a:2:{s:8:\"tag_name\";N;s:6:\"result\";s:1:\"1\";}i:2;a:2:{s:8:\"tag_name\";s:2:\"TV\";s:6:\"result\";s:1:\"1\";}}";

    let actual: Vec<TagItem> = from_str(raw).expect("deserialize php serialized vec");
    assert_eq!(
      actual,
      vec![
        TagItem {
          tag_name: Some("动画".to_string()),
          result: "2".to_string(),
        },
        TagItem {
          tag_name: None,
          result: "1".to_string(),
        },
        TagItem {
          tag_name: Some("TV".to_string()),
          result: "1".to_string(),
        },
      ]
    );
  }

  #[test]
  fn deserialize_scalar_string() {
    let raw = "s:6:\"动画\";";
    let value: String = from_str(raw).expect("deserialize php string");
    assert_eq!(value, "动画");
  }
}
