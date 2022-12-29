use log::error;
use poise::{serenity_prelude::{CreateMessage, ChannelId, CreateEmbed}, CreateReply};
use rand::{distributions::{Alphanumeric, DistString}};

use crate::{Context, Error};

/// Command reference
#[poise::command(slash_command, prefix_command, guild_only)]
pub async fn cmdref(
    ctx: Context<'_>,
) -> Result<(), Error> {
    ctx.say(
        "
- **newhook:** newhook <comment>

EX: ``newhook 'My new webhook'``

- **delhook:** delhook <webhook id>

EX: ``delhook ID``

- **list:** list

EX: ``list``

- **newrepo:** newrepo <webhook id> <repo org/owner> <repo name> <channel_id>

EX: ``newrepo ID Git-Logs MyRepo #github``

- **delrepo:** delrepo <repo id>

EX: ``delrepo ID``

- **setrepoevents:** setrepoevents <repo id> <space seperated list of events>

EX: ``setrepoevents ID 'push pull_request'``

- **setrepochannel:** delrepoevents <repo id>

EX: ``delrepoevents ID``

- **setrepochannel:** setrepochannel <repo id> <channel_id>

EX: ``setrepochannel ID #github``
        "
    ).await?;

    Ok(())
}

/// Lsts all webhooks in a guild with their respective repos and channel IDs
#[poise::command(slash_command, prefix_command, guild_only, required_permissions = "MANAGE_GUILD")]
pub async fn list(
    ctx: Context<'_>,
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
        // If it doesn't, return an error
        sqlx::query!(
            "INSERT INTO guilds (guild_id) VALUES ($1)",
            ctx.guild_id().unwrap().to_string()
        )
        .execute(&data.pool)
        .await?;

        ctx.say("This guild doesn't have any webhooks yet. Get started with ``/newhook`` (or ``git!newhook``)").await?;

        return Ok(());
    } else {
        // Get all webhooks
        let webhooks = sqlx::query!(
            "SELECT id, guild_id, comment, created_at FROM webhooks WHERE guild_id = $1",
            ctx.guild_id().unwrap().to_string()
        )
        .fetch_all(&data.pool)
        .await;

        match webhooks {
            Ok(webhooks) => {
                let mut embeds = Vec::new();

                for webhook in webhooks {
                    // Get repos of webhook
                    let repos = sqlx::query!(
                        "SELECT id, repo_name, channel_id, events FROM repos WHERE webhook_id = $1",
                        webhook.id
                    )
                    .fetch_all(&data.pool)
                    .await?;

                    let mut repo_str = String::new();

                    for repo in repos {
                        let mut event_whitelist = repo.events.join(", ");

                        if event_whitelist.is_empty() {
                            event_whitelist = "All events allowed".to_string();
                        }    

                        repo_str.push_str(&format!(
                            "__**{repo_name}**__\n\n*Channel ID:* {channel_id}\n*Repo ID:* {id}\n*Repo Name:* {repo_name}\n*Events Whitelist:* {event_whitelist}\n\n",
                            repo_name = repo.repo_name,
                            channel_id = repo.channel_id,
                            id = repo.id,
                            event_whitelist = event_whitelist
                        ));
                    }

                    embeds.push(
                        CreateEmbed::new()
                        .title(format!("Webhook \"{}\"", webhook.comment))
                        .field("ID", webhook.id, false)
                        .field("Created at", webhook.created_at.to_string(), false)
                        .field("Repos", repo_str, false)
                    );
                };

                let mut cr = CreateReply::default()
                .content("Here are all the webhooks in this guild:");

                cr.embeds = embeds;

                ctx.send(cr).await?;

                return Ok(());
            },
            Err(e) => {
                error!("Error fetching webhooks: {:?}", e);
                ctx.say("This guild doesn't have any webhooks yet. Get started with ``/newhook`` (or ``git!newhook``)").await?;

                return Ok(());
            }
        }
    }


}

/// Creates a new webhook in a guild for sending Github notifications
#[poise::command(slash_command, prefix_command, guild_only, guild_cooldown = 60, required_permissions = "MANAGE_GUILD")]
pub async fn newhook(
    ctx: Context<'_>,
    #[description = "The comment for the webhook"] comment: String,
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
        // If it doesn't, create it
        sqlx::query!(
            "INSERT INTO guilds (guild_id) VALUES ($1)",
            ctx.guild_id().unwrap().to_string()
        )
        .execute(&data.pool)
        .await?;
    }

    // Check webhook count
    let webhook_count = sqlx::query!(
        "SELECT COUNT(1) FROM webhooks WHERE guild_id = $1",
        ctx.guild_id().unwrap().to_string()
    )
    .fetch_one(&data.pool)
    .await?;

    if webhook_count.count.unwrap_or_default() >= 3 {
        ctx.say("You can't have more than 3 webhooks per guild").await?;
        return Ok(());
    }

    // Create the webhook
    let id = Alphanumeric.sample_string(&mut rand::thread_rng(), 32);

    let webh_secret = Alphanumeric.sample_string(&mut rand::thread_rng(), 128);

    sqlx::query!(
        "INSERT INTO webhooks (id, guild_id, comment, secret) VALUES ($1, $2, $3, $4)",
        id,
        ctx.guild_id().unwrap().to_string(),
        comment,
        webh_secret
    )
    .execute(&data.pool)
    .await?;

    ctx.say("Webhook created! Trying to DM you the credentials...").await?;

    // Create a new dm channel with the user if not slash command
    let dm_channel = ctx.author().create_dm_channel(&ctx).await;

    let dm = match dm_channel {
        Ok(dm) => dm,
        Err(_) => {
            ctx.say("I couldn't create a DM channel with you, please enable DMs from server members").await?;
            return Ok(());
        }
    };

    dm.id.send_message(
        &ctx,
        CreateMessage::new()
        .content(
            format!(
                "
Next, add the following webhook to your Github repositories (or organizations): `{respond_url}/kittycat?id={id}`

Set the `Secret` field to `{webh_secret}` and ensure that Content Type is set to `application/json`. 

When creating repositories, use `{id}` as the ID.
            
**Note that the above URL and secret is unique and should not be shared with others**

**Delete this message after you're done!**
                ",
                respond_url=std::env::var("RESPOND_URL").unwrap(),
                id=id,
                webh_secret=webh_secret
            )    
        )
    ).await?;

    ctx.say("Webhook created! Check your DMs for the webhook information.").await?;
    
    Ok(())
}

/// Creates a new repository for a webhook
#[poise::command(slash_command, prefix_command, guild_only, guild_cooldown = 60, required_permissions = "MANAGE_GUILD")]
pub async fn newrepo(
    ctx: Context<'_>,
    #[description = "The webhook ID to use"] webhook_id: String,
    #[description = "The repo owner or organization"] owner: String,
    #[description = "The repo name"] name: String,
    #[description = "The channel to send to"] channel: ChannelId,
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

    // Check webhook count
    let webhook_count = sqlx::query!(
        "SELECT COUNT(1) FROM webhooks WHERE guild_id = $1",
        ctx.guild_id().unwrap().to_string()
    )
    .fetch_one(&data.pool)
    .await?;

    let count = webhook_count.count.unwrap_or_default();

    if count == 0 {
        return Err("You don't have any webhooks in this guild! Use ``/newhook`` (or ``git!newhook``) to create one".into());
    } else {
        // Check if the webhook exists
        let webhook = sqlx::query!(
            "SELECT COUNT(1) FROM webhooks WHERE id = $1 AND guild_id = $2",
            webhook_id,
            ctx.guild_id().unwrap().to_string()
        )
        .fetch_one(&data.pool)
        .await?;

        if webhook.count.unwrap_or_default() == 0 {
            return Err("That webhook doesn't exist! Use ``/newhook`` (or ``git!newhook``) to create one".into());
        }

        // Check if the repo exists
        let repo = sqlx::query!(
            "SELECT COUNT(1) FROM repos WHERE id = $1 AND webhook_id = $2",
            name,
            webhook_id
        )
        .fetch_one(&data.pool)
        .await?;

        if repo.count.unwrap_or_default() == 0 {
            // If it doesn't, create it
            let id = Alphanumeric.sample_string(&mut rand::thread_rng(), 32);

            sqlx::query!(
                "INSERT INTO repos (id, webhook_id, repo_name, channel_id, guild_id) VALUES ($1, $2, $3, $4, $5)",
                id,
                webhook_id,
                (owner+"/"+&name).to_lowercase(),
                channel.to_string(),
                ctx.guild_id().unwrap().to_string()
            )
            .execute(&data.pool)
            .await?;

            ctx.say(
                format!("Repository created with ID of ``{id}``!", id=id)
            ).await?;

            Ok(())
        } else {
            return Err("That repo already exists! Use ``/delrepo`` (or ``git!delrepo``) to delete it".into());
        }
    }
}

/// Deletes a webhook
#[poise::command(slash_command, prefix_command, guild_only, guild_cooldown = 60, required_permissions = "MANAGE_GUILD")]
pub async fn delhook(
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

    sqlx::query!(
        "DELETE FROM webhooks WHERE id = $1 AND guild_id = $2",
        id,
        ctx.guild_id().unwrap().to_string()
    )
    .execute(&data.pool)
    .await?;

    ctx.say("Webhook deleted if it exists!").await?;

    Ok(())
}

/// Deletes a repository
#[poise::command(slash_command, prefix_command, guild_only, guild_cooldown = 60, required_permissions = "MANAGE_GUILD")]
pub async fn delrepo(
    ctx: Context<'_>,
    #[description = "The repo ID"] id: String,
) -> Result<(), Error> { 
    let data = ctx.data();

    sqlx::query!(
        "DELETE FROM repos WHERE id = $1 AND guild_id = $2",
        id,
        ctx.guild_id().unwrap().to_string()
    )
    .execute(&data.pool)
    .await?;

    ctx.say("Repo deleted!").await?;

    Ok(())
}

/// Sets a event whitelist for a repository
#[poise::command(slash_command, prefix_command, guild_only, guild_cooldown = 60, required_permissions = "MANAGE_GUILD")]
pub async fn setrepoevents(
    ctx: Context<'_>,
    #[description = "The repo ID"] id: String,
    #[description = "The events to whitelist"] events: String,
) -> Result<(), Error> { 
    let data = ctx.data();

    // Check if the repo exists
    let repo = sqlx::query!(
        "SELECT COUNT(1) FROM repos WHERE id = $1 AND guild_id = $2",
        id,
        ctx.guild_id().unwrap().to_string()
    )
    .fetch_one(&data.pool)
    .await?;

    if repo.count.unwrap_or_default() == 0 {
        return Err("That repo doesn't exist! Use ``/newrepo`` (or ``git!newrepo``) to create one".into());
    }

    if events.contains(",") {
        return Err("Events are space seperated, not comma seperated!".into());
    }

    let events_vec = events.split(" ").map(|s| s.to_string()).collect::<Vec<String>>();

    sqlx::query!(
        "UPDATE repos SET events = $1 WHERE id = $2 AND guild_id = $3",
        &events_vec,
        id,
        ctx.guild_id().unwrap().to_string()
    )
    .execute(&data.pool)
    .await?;


    Ok(())
}

/// Deletes a event whitelist for a repository
#[poise::command(slash_command, prefix_command, guild_only, guild_cooldown = 60, required_permissions = "MANAGE_GUILD")]
pub async fn delrepoevents(
    ctx: Context<'_>,
    #[description = "The repo ID"] id: String,
) -> Result<(), Error> { 
    let data = ctx.data();

    // Check if the repo exists
    let repo = sqlx::query!(
        "SELECT COUNT(1) FROM repos WHERE id = $1 AND guild_id = $2",
        id,
        ctx.guild_id().unwrap().to_string()
    )
    .fetch_one(&data.pool)
    .await?;

    if repo.count.unwrap_or_default() == 0 {
        return Err("That repo doesn't exist! Use ``/newrepo`` (or ``git!newrepo``) to create one".into());
    }

    sqlx::query!(
        "UPDATE repos SET events = $1 WHERE id = $2 AND guild_id = $3",
        &vec![],
        id,
        ctx.guild_id().unwrap().to_string()
    )
    .execute(&data.pool)
    .await?;


    Ok(())
}

/// Sets a event whitelist for a repository
#[poise::command(slash_command, prefix_command, guild_only, guild_cooldown = 60, required_permissions = "MANAGE_GUILD")]
pub async fn setrepochannel(
    ctx: Context<'_>,
    #[description = "The repo ID"] id: String,
    #[description = "The new channel ID"] channel: ChannelId,
) -> Result<(), Error> { 
    let data = ctx.data();

    // Check if the repo exists
    let repo = sqlx::query!(
        "SELECT COUNT(1) FROM repos WHERE id = $1 AND guild_id = $2",
        id,
        ctx.guild_id().unwrap().to_string()
    )
    .fetch_one(&data.pool)
    .await?;

    if repo.count.unwrap_or_default() == 0 {
        return Err("That repo doesn't exist! Use ``/newrepo`` (or ``git!newrepo``) to create one".into());
    }

    sqlx::query!(
        "UPDATE repos SET channel_id = $1 WHERE id = $2 AND guild_id = $3",
        channel.to_string(),
        id,
        ctx.guild_id().unwrap().to_string()
    )
    .execute(&data.pool)
    .await?;


    Ok(())
}
