use poise::{CreateReply, serenity_prelude::CreateAttachment};
use serde::{Serialize, Deserialize};

use crate::{Context, Error};

#[derive(Serialize, Deserialize)]
struct RepoBackup {
    protocol: u32,
    repos: Vec<Repo>,
}

#[derive(Serialize, Deserialize)]
struct Repo {
    repo_name: String,
    channel_id: String,
    events: Vec<String>,
}

/// Backups the repositories of a webhook to a JSON file
#[poise::command(slash_command, prefix_command, guild_only, required_permissions = "MANAGE_GUILD")]
pub async fn backuprepos(
    ctx: Context<'_>,
    #[description = "The webhook ID"] id: String,
) -> Result<(), Error> {
    let data = ctx.data();

    // Check if the guild exists on our DB
    let guild = sqlx::query!(
        "SELECT COUNT(1) FROM guilds WHERE guild_id = $1",
        ctx.guild_id().unwrap().to_string()
    )
    .fetch_one(&data.pool)
    .await?;
    
    if guild.count.unwrap_or_default() == 0 {
        // If it doesn't, return a error
        return Err("You don't have any webhooks in this guild! Use ``/newhook`` (or ``git!newhook``) to create one".into());
    }

    // Check if the webhook exists
    let webhook = sqlx::query!(
        "SELECT COUNT(1) FROM webhooks WHERE id = $1 AND guild_id = $2",
        id,
        ctx.guild_id().unwrap().to_string()
    )
    .fetch_one(&data.pool)
    .await?;

    if webhook.count.unwrap_or_default() == 0 {
        return Err("That webhook doesn't exist! Use ``/newhook`` (or ``git!newhook``) to create one".into());
    }

    let rows = sqlx::query!(
        "SELECT repo_name, channel_id, events FROM repos WHERE webhook_id = $1",
        id
    )
    .fetch_all(&data.pool)
    .await?;

    let mut repos = Vec::new();

    for row in rows {
        repos.push(Repo {
            repo_name: row.repo_name,
            channel_id: row.channel_id,
            events: row.events,
        });
    }

    let backup = RepoBackup {
        protocol: 1,
        repos,
    };

    let json = serde_json::to_string(&backup)?;

    let msg = CreateReply::new()
    .content("Here's your backup file!")
    .attachment(
        CreateAttachment::bytes(json.as_bytes(), "repos_".to_string() + &id + ".json")
    );

    ctx.send(msg).await?;

    Ok(())
}