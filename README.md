# Github-v2

Github v2 is the rewrite of the original Github webhook logger.

- bot -> the frontend bot that allows configuration of the webhook logger
- webserver -> the webserver that hosts the webhook logger

## Self Hosting

### Compiling

Fill out ``bot/.env`` and ``webserver/.env`` (see the ``.env.sample`` files there)

- Run ``make selfhost`` to build the bot.
- Enter the ``webserver`` folder and run ``make`` to build the webserver.

### Running

You should ideally make this 2 systemd services in production.

- Run the bot with ``make run``.