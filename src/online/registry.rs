use serde::{Serialize, Deserialize};
use crate::core::*;
use std::sync::Arc;

#[cfg(target_os="windows")]
static OS_NAME: &str = "windows";

#[cfg(target_os="linux")]
static OS_NAME: &str = "linux";

#[cfg(target_os="macos")]
static OS_NAME: &str = "darwin";

#[async_trait::async_trait]
pub trait Registry : Send + Sync + From<Arc<Config>> {
    async fn get_entries(&self) -> Result<Vec<String>, Error>;
    async fn get_entry(&self, id: &str) -> Result<Entry, Error>;
}



#[derive(Debug, Serialize, Deserialize, Clone, Default)]
pub struct Entry {
    pub name: String,
    pub description: String,
    #[serde(default)]
    pub configs: Vec<EntryConfig>,
}

#[derive(Debug, Serialize, Deserialize, Clone, Default)]
pub struct EntryConfig {
    pub platform: String,
    #[serde(default)]
    pub app: Option<EntryApp>,
    #[serde(default)]
    pub service: Option<EntryService>
}

impl EntryConfig {
    pub fn is_compatible(&self) -> bool {
        self.platform == "any" || self.platform == OS_NAME
    }
}

#[derive(Debug, Serialize, Deserialize, Clone, Default)]
pub struct EntryApp {
    pub name: String,
    pub command: String,
    #[serde(default)]
    pub args: Vec<String>,
    #[serde(default)]
    pub environment: Vec<String>
}

impl Into<App> for EntryApp {
    fn into(self) -> App {
        App::builder()
            .with_name(&self.name)
            .with_command(&self.command)
            .with_args(self.args.iter().map(|s| s.as_str()).collect())
            .with_environment(self.environment.iter().map(|s| s.as_str()).collect())
            .into()
    }
}

#[derive(Debug, Serialize, Deserialize, Clone, Default)]
pub struct EntryService {
    pub domain: String,
    pub website: String,
    #[serde(rename="httpUrl")]
    pub http_url: String,
    #[serde(rename="gitUrl")]
    pub git_url: String,
    pub pattern: String    
}

impl Into<Service> for EntryService {
    fn into(self) -> Service {
        Service::builder()
            .with_domain(&self.domain)
            .with_website(&self.website)
            .with_git_url(&self.git_url)
            .with_http_url(&self.http_url)
            .with_pattern(&self.pattern)
            .into()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn is_compatible() {
        assert_eq!(EntryConfig{
            platform: "any".to_string(),
            ..Default::default()
        }.is_compatible(), true);
        
        assert_eq!(EntryConfig{
            platform: "windows".to_string(),
            ..Default::default()
        }.is_compatible(), OS_NAME == "windows");
        
        assert_eq!(EntryConfig{
            platform: "linux".to_string(),
            ..Default::default()
        }.is_compatible(), OS_NAME == "linux");
        
        assert_eq!(EntryConfig{
            platform: "darwin".to_string(),
            ..Default::default()
        }.is_compatible(), OS_NAME == "darwin");
    }
}