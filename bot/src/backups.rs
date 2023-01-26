use poise::{serenity_prelude::{CreateMessage, ChannelId, CreateEmbed}, CreateReply};

/// Backups the repositories of a webhook to a JSON file
#[poise::command(slash_command, prefix_command, guild_only, required_permissions = "MANAGE_GUILD")]
pub async fn backuprepos() -> Result<(), Error> {
    
}