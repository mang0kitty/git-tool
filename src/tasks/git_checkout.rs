use super::{core, Task};
use crate::{errors, core::Target};
use tokio::process::Command;

pub struct GitCheckout {
    pub branch: String,
}

#[async_trait::async_trait]
impl Task for GitCheckout {
    async fn apply_repo(&self, repo: &core::Repo) -> Result<(), core::Error> {
        let status = Command::new("git")
            .args(vec!["checkout", "-B", &self.branch])
            .current_dir(repo.get_path())
            .status().await?;

        match status.code() {
            Some(0) => Ok(()),
            Some(_) => Err(errors::system("The git init command failed with a non-zero exit code.", "Please check the output printed by Git to determine why the command failed and take appropriate action.")),
            None => Ok(())
        }
    }

    async fn apply_scratchpad(&self, _scratch: &core::Scratchpad) -> Result<(), core::Error> {
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::tasks::GitInit;

    #[tokio::test]
    async fn test_repo() {
        let repo = core::Repo::new(
            "github.com/sierrasoftworks/test2", 
            get_dev_dir().join("github.com").join("sierrasoftworks").join("test2"));

        assert_eq!(repo.get_path().exists(), true);

        let result = sequence![
            GitInit{},
            GitCheckout{
                branch: "test".into()
            }
        ].apply_repo(&repo).await;

        if result.is_err() {
            std::fs::remove_dir_all(repo.get_path().join(".git")).unwrap();
            result.unwrap();
        }
    }

    #[tokio::test]
    async fn test_scratch() {
        let scratch = core::Scratchpad::new(
            "2019w15", 
            get_dev_dir().join("scratch").join("2019w15"));

        let task = GitCheckout{
            branch: "test".into(),
        };
        assert_eq!(scratch.get_path().exists(), true);

        task.apply_scratchpad(&scratch).await.unwrap();
        assert_eq!(scratch.get_path().join(".git").exists(), false);
    }

    fn get_dev_dir() -> std::path::PathBuf {
        std::path::PathBuf::from(file!())
            .parent()
            .and_then(|f| f.parent())
            .and_then(|f| f.parent())
            .and_then(|f| Some(f.join("test")))
            .and_then(|f| Some(f.join("devdir")))
            .unwrap()
    }
}