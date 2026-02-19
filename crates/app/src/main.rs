use anyhow::Context;
use clap::{Parser, Subcommand};

mod worker;

#[derive(Debug, Parser)]
#[command(name = "app")]
struct Cli {
  #[command(subcommand)]
  command: Commands,
}

#[derive(Debug, Subcommand)]
enum Commands {
  Worker {
    #[command(subcommand)]
    command: WorkerCommands,
  },
  Server,
}

#[derive(Debug, Subcommand)]
enum WorkerCommands {
  Canal,
  Timeline,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
  common::init_tracing();

  let cli = Cli::parse();

  match cli.command {
    Commands::Worker { command } => match command {
      WorkerCommands::Canal => worker::canal::run().await,
      WorkerCommands::Timeline => worker::timeline::run().await,
    },
    Commands::Server => run_server().await,
  }
}

async fn run_server() -> anyhow::Result<()> {
  tracing::info!("server subcommand placeholder is ready");
  tokio::signal::ctrl_c().await.context("wait ctrl-c")?;
  tracing::info!("server placeholder shutdown");
  Ok(())
}
