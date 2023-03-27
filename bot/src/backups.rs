use poise::{CreateReply, serenity_prelude::{CreateAttachment, Attachment}};
use serde::{Serialize, Deserialize};

use crate::{Context, Error};

#[derive(Serialize, Deserialize)]
struct Repo {
    repo_name: String,
    channel_id: String,
    events: Vec<String>,
}

/// Backups the repositories of a webhook to a JSON file
#[poise::command(category = "Backups", slash_command, prefix_command, guild_only, required_permissions = "MANAGE_GUILD")]
pub async fn backup(
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

    let json = serde_json::to_string(&repos)?;

    let msg = CreateReply::new()
    .content("Here's your backup file!")
    .attachment(
        CreateAttachment::bytes(json.as_bytes(), "repos_".to_string() + &id + ".glb")
    );

    ctx.send(msg).await?;

    Ok(())
}

/// Restore a created backup to a webhook
#[poise::command(category = "Backups", slash_command, prefix_command, guild_only, required_permissions = "MANAGE_GUILD")]
pub async fn restore(
    ctx: Context<'_>,
    #[description = "The webhook ID to restore the backup to"] id: String,
    #[description = "The backup file"] file: Attachment,
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

    let backup = file.download().await?;

    let repos: Vec<Repo> = serde_json::from_slice(&backup)?;

    let mut inserted = 0;
    let mut updated = 0;

    for repo in repos {
        // Check that the repo exists
        let repo_exists = sqlx::query!(
            "SELECT COUNT(1) FROM repos WHERE lower(repo_name) = $1 AND webhook_id = $2",
            repo.repo_name.to_lowercase(),
            id
        )
        .fetch_one(&data.pool)
        .await?;

        if repo_exists.count.unwrap_or_default() == 0 {
            // If it doesn't, create it
            sqlx::query!(
                "INSERT INTO repos (repo_name, webhook_id, channel_id, events) VALUES ($1, $2, $3, $4)",
                repo.repo_name,
                id,
                repo.channel_id,
                &repo.events
            )
            .execute(&data.pool)
            .await?;

            inserted += 1;
        } else {
            // If it does, update it
            sqlx::query!(
                "UPDATE repos SET channel_id = $1, events = $2 WHERE lower(repo_name) = $3 AND webhook_id = $4",
                repo.channel_id,
                &repo.events,
                repo.repo_name.to_lowercase(),
                id
            )
            .execute(&data.pool)
            .await?;

            updated += 1;
        }
    }

    ctx.say(
        format!(r#"
**Summary**

- **Inserted:** {}
- **Updated:** {}"#, inserted, updated)
    ).await?;

    Ok(())
}
