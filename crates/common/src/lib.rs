pub mod config;

#[derive(Debug)]
pub struct ErrorAt {
  pub file: &'static str,
  pub line: u32,
  pub message: &'static str,
  pub source: anyhow::Error,
}

impl std::fmt::Display for ErrorAt {
  fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
    write!(f, "{}", self.message)
  }
}

impl std::error::Error for ErrorAt {
  fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
    Some(self.source.as_ref())
  }
}

pub fn locate_error(err: &anyhow::Error) -> Option<&ErrorAt> {
  for cause in err.chain() {
    if let Some(at) = cause.downcast_ref::<ErrorAt>() {
      return Some(at);
    }
  }
  None
}

pub trait ResultExt<T> {
  #[track_caller]
  fn context_loc(self, message: &'static str) -> anyhow::Result<T>;
}

impl<T, E> ResultExt<T> for Result<T, E>
where
  E: Into<anyhow::Error>,
{
  #[track_caller]
  fn context_loc(self, message: &'static str) -> anyhow::Result<T> {
    self.map_err(|err| error_at(err, message))
  }
}

#[cold]
#[inline(never)]
#[track_caller]
fn error_at<E>(err: E, message: &'static str) -> anyhow::Error
where
  E: Into<anyhow::Error>,
{
  let caller = std::panic::Location::caller();
  anyhow::Error::new(ErrorAt {
    file: caller.file(),
    line: caller.line(),
    message,
    source: err.into(),
  })
}

pub fn init_tracing() {
  let filter = std::env::var("RUST_LOG").unwrap_or_else(|_| "info".to_string());
  let use_json = std::env::var("RUST_LOG_JSON")
    .map(|v| matches!(v.as_str(), "1" | "true" | "TRUE" | "True"))
    .unwrap_or(false);

  let builder = tracing_subscriber::fmt()
    .with_env_filter(filter)
    .with_target(true)
    .with_file(true)
    .with_line_number(true);

  if use_json {
    let _ = builder.json().try_init();
  } else {
    let _ = builder.try_init();
  }
}
