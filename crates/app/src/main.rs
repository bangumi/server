use anyhow::Context;
use app::{server, worker};
use clap::{Parser, Subcommand};

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
  server::run().await.context("run server")
}
