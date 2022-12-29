# Git Logs

Rewrite of the original Github webhook logger (Git Logs).

---

## Supported Events

See [here](https://docs.github.com/en/developers/webhooks-and-events/webhook-events-and-payloads) for a list of all events.

- `push`
- `star`
- `issues`
- `pull_request`
- `issue_comment`
- `pull_request_review_comment`
- `create`
- `check_suite`
- `status`
- `release`
- `commit_comment`
- `deployment`
- `deployment_status`
- `workflow_run`
- `dependabot`

**More coming soon**

---

## The Stack

- bot -> the frontend bot that allows configuration of the webhook logger
- webserver -> the webserver that hosts the webhook logger

---

## Self Hosting

### Database

Run the following in ``psql`` to create the database:

```sql
CREATE DATABASE github;
\c github
\i schema.sql
```

### Compiling

Fill out ``bot/.env`` and ``webserver/.env`` (see the ``.env.sample`` files there)

- Run ``make selfhost`` to build the bot.
- Enter the ``webserver`` folder and run ``make`` to build the webserver.

### Running

You should ideally make this 2 systemd services in production.

- Run the bot with ``make run`` (in the ``bot`` folder).
- Run the webserver with ``./webserver`` (in the ``webserver`` folder).

**Note that a ``206`` status code is returned if ``repo_url`` is not added to the webhook**

---

## License

This project is licensed under the MIT License
