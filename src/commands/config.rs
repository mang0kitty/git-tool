use clap::{App, SubCommand, ArgMatches, Arg};
use super::Command;
use super::*;
use super::async_trait;
use online::registry::Registry;

pub struct ConfigCommand {

}

impl Command for ConfigCommand {
    fn name(&self) -> String {
        String::from("config")
    }

    fn app<'a, 'b>(&self) -> App<'a, 'b> {
        SubCommand::with_name(self.name().as_str())
            .version("1.0")
            .about("manage your Git-Tool configuration file")
            .help_message("This tool allows you to easily make changes to your Git-Tool config file.")

            .subcommand(SubCommand::with_name("list")
                .version("1.0")
                .alias("ls")
                .about("list available config templates")
                .help_message("Gets the list of config templates which are available through the Git-Tool registry."))

            .subcommand(SubCommand::with_name("add")
                .version("1.0")
                .about("adds a configuration template to your current config file")
                .help_message("Adds a configuration template from the Git-Tool online registry to your config file.")
                .arg(Arg::with_name("config")
                    .short("c")
                    .long("config")
                    .help("the configuration file to update")
                    .env("GITTOOL_CONFIG")
                    .takes_value(true))
                .arg(Arg::with_name("id")
                    .index(1)
                    .help("the id of the configuration template you want to add")
                    .required(true)))
    }
}
    
#[async_trait]
impl<F: FileSource, L: Launcher, R: Resolver> CommandRunnable<F, L, R> for ConfigCommand {
    async fn run<'a>(&self, core: &core::Core<F, L, R>, matches: &ArgMatches<'a>) -> Result<i32, errors::Error> {
        match matches.subcommand() {
            ("list", Some(_args)) => {
                let registry = crate::online::GitHubRegistry::from(core.config.clone());

                let entries = registry.get_entries().await?;
                for entry in entries {
                    println!("{}", entry);
                }
            },
            ("add", Some(args)) => {
                let id = args.value_of("id").ok_or(errors::user(
                    "You have not provided an ID for the config template you wish to add.",
                    ""))?;

                let registry = crate::online::GitHubRegistry::from(core.config.clone());
                let entry = registry.get_entry(id).await?;

                println!("Applying {}", entry.name);
                println!("{}", entry.description);

                let mut cfg = core.config.clone();
                for ec in entry.configs {
                    if ec.is_compatible() {
                        cfg = Arc::new(cfg.add(ec));
                    }
                }

                match matches.value_of("config") {
                    Some(path) => {
                        core.file_source.write(&std::path::PathBuf::from(path), cfg.to_string()?).await?;
                    },
                    None => {
                        println!("{}", cfg.to_string()?);
                    }
                }
                
            },
            _ => {
                println!("{}", core.config.to_string()?);
            }
        }

        Ok(0)
    }

    async fn complete<'a>(&self, core: &Core<F, L, R>, completer: &Completer, matches: &ArgMatches<'a>) {
        match matches.subcommand() {
            ("list", _) => {

            },
            ("add", _) => {
                match online::GitHubRegistry::from(core.config.clone()).get_entries().await {
                    Ok(entries) => {
                        completer.offer_many(entries);
                    },
                    _ => {}
                }
            },
            _ => {
                completer.offer_many(vec!["list", "add"]);
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::core::{Config};
    use clap::ArgMatches;

    #[tokio::test]
    async fn run() {
        let args = ArgMatches::default();
        let cfg = Config::from_str("directory: /dev").unwrap();
        let core = Core::builder().with_config(&cfg).build();

        let cmd = ConfigCommand{};

        match cmd.run(&core, &args).await {
            Ok(_) => {},
            Err(err) => {
                panic!(err.message())
            }
        }
    }
}