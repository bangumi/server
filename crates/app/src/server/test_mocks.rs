pub(super) struct MockPool {
  pub subject_image_repo: super::subjects::MockSubjectImageRepo,
  pub character_image_repo: super::characters::MockCharacterImageRepo,
  pub person_image_repo: super::persons::MockPersonImageRepo,
}

impl MockPool {
  pub(super) fn new() -> Self {
    Self {
      subject_image_repo: super::subjects::MockSubjectImageRepo::new(),
      character_image_repo: super::characters::MockCharacterImageRepo::new(),
      person_image_repo: super::persons::MockPersonImageRepo::new(),
    }
  }
}
