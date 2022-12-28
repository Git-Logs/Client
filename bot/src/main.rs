use std::{time::Duration};

use dotenv::dotenv;
use log::{error, info};
use poise::serenity_prelude::{
    self as serenity, FullEvent, UserId,
};
use sqlx::postgres::PgPoolOptions;

mod help;
mod core;

pub const VERSION: &str = env!("CARGO_PKG_VERSION");

type Error = Box<dyn std::error::Error + Send + Sync>;
type Context<'a> = poise::Context<'a, Data, Error>;
// User data, which is stored and accessible in all command invocations
pub struct Data {
    pub pool: sqlx::PgPool,
}

#[poise::command(prefix_command, hide_in_help)]
async fn register(ctx: Context<'_>) -> Result<(), Error> {
    poise::builtins::register_application_commands_buttons(ctx).await?;
    Ok(())
}

async fn on_error(error: poise::FrameworkError<'_, Data, Error>) {
    // This is our custom error handler
    // They are many errors that can occur, so we only handle the ones we want to customize
    // and forward the rest to the default handler
    match error {
        poise::FrameworkError::Setup { error, .. } => panic!("Failed to start bot: {:?}", error),
        poise::FrameworkError::Command { error, ctx } => {
            error!("Error in command `{}`: {:?}", ctx.command().name, error,);
            ctx.say(format!(
                "There was an error running this command: {}",
                error
            ))
            .await
            .unwrap();
        }
        poise::FrameworkError::CommandCheckFailed { error, ctx } => {
            error!(
                "[Possible] error in command `{}`: {:?}",
                ctx.command().name,
                error,
            );
            if let Some(error) = error {
                error!("Error in command `{}`: {:?}", ctx.command().name, error,);
                ctx.say(format!(
                    "Whoa there, do you have permission to do this?: {}",
                    error
                ))
                .await
                .unwrap();
            } else {
                ctx.say("You don't have permission to do this but we couldn't figure out why...")
                    .await
                    .unwrap();
            }
        }
        error => {
            if let Err(e) = poise::builtins::on_error(error).await {
                error!("Error while handling error: {}", e);
            }
        }
    }
}

async fn event_listener(event: &FullEvent, _user_data: &Data) -> Result<(), Error> {
    match event {
        FullEvent::InteractionCreate { interaction, ctx: _ } => {
            info!("Interaction received: {:?}", interaction.id());
        }
        FullEvent::Ready {
            data_about_bot,
            ctx: _,
        } => {
            // Always wait a bit here for cache to finish up
            tokio::time::sleep(Duration::from_secs(2)).await;

            info!(
                "{} is ready!",
                data_about_bot.user.name
            );
        }
        _ => {}
    }

    Ok(())
}

#[tokio::main]
async fn main() {
    const MAX_CONNECTIONS: u32 = 3; // max connections to the database, we don't need too many here

    dotenv().ok();

    env_logger::init();

    let http =
        serenity::HttpBuilder::new(std::env::var("DISCORD_TOKEN").expect("missing DISCORD_TOKEN")).build();

    let client_builder =
        serenity::ClientBuilder::new_with_http(
            http, 
            serenity::GatewayIntents::MESSAGE_CONTENT | serenity::GatewayIntents::GUILD_MESSAGES | serenity::GatewayIntents::GUILDS
        );

    let owners = std::env::var("OWNERS")
        .expect("missing OWNERS")
        .split(',')
        .map(|x| x.parse::<UserId>().expect("invalid owner"))
        .collect::<Vec<_>>();

    // Convert to hashset of HashSet<UserId>
    let owners = owners.into_iter().collect::<std::collections::HashSet<_>>();

    let framework = poise::Framework::new(
        poise::FrameworkOptions {
            owners,
            initialize_owners: false,
            prefix_options: poise::PrefixFrameworkOptions {
                prefix: Some("git!".into()),
                ..poise::PrefixFrameworkOptions::default()
            },
            listener: |event, _ctx, user_data| Box::pin(event_listener(event, user_data)),
            commands: vec![
                register(),
                help::simplehelp(),
                help::help(),
                core::list(),
                core::newhook(),
                core::newrepo(),
                core::delhook(),
                core::delrepo(),
            ],
            /// This code is run before every command
            pre_command: |ctx| {
                Box::pin(async move {
                    info!(
                        "Executing command {} for user {} ({})...",
                        ctx.command().qualified_name,
                        ctx.author().name,
                        ctx.author().id
                    );
                })
            },
            /// This code is run after every command returns Ok
            post_command: |ctx| {
                Box::pin(async move {
                    info!(
                        "Done executing command {} for user {} ({})...",
                        ctx.command().qualified_name,
                        ctx.author().name,
                        ctx.author().id
                    );
                })
            },
            on_error: |error| Box::pin(on_error(error)),
            ..Default::default()
        },
        move |_ctx, _ready, _framework| {
            Box::pin(async move {
                Ok(Data {
                    pool: PgPoolOptions::new()
                        .max_connections(MAX_CONNECTIONS)
                        .connect(&std::env::var("DATABASE_URL").expect("missing DATABASE_URL"))
                        .await
                        .expect("Could not initialize connection"),
                })
            })
        },
    );

    let mut client = client_builder
        .framework(framework)
        .await
        .expect("Error creating client");

    if let Err(why) = client.start().await {
        error!("Client error: {:?}", why);
    }
}