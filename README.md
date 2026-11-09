mgw-core-manager
=======

![Image](https://img.shields.io/github/v/tag/SENERGY-Platform/mgw-core-manager?filter=v%2A&label=release)

![Image](https://img.shields.io/github/v/tag/SENERGY-Platform/mgw-core-manager?filter=lib%2A&label=latest)

![Image](https://img.shields.io/github/v/tag/SENERGY-Platform/mgw-core-manager?filter=client%2A&label=latest)

Generate Swagger Docs:

    swag init -g routes.go -o handler/http_hdl/swagger_docs -dir handler/http_hdl/standard,handler/http_hdl/shared --parseDependency --instanceName standard
    swag init -g routes.go -o handler/http_hdl/swagger_docs -dir handler/http_hdl/restricted,handler/http_hdl/shared --parseDependency --instanceName restricted
## Core containers

The core containers have restart policy "no", so docker never starts them on its own. The core-manager owns
their lifecycle; the installer only creates them (`docker compose up --no-start`).

- **Start:** when the core-manager starts, it waits until every socket in `containers.wait_sockets` accepts
  connections, then runs `docker compose -f <compose_file_path> start`. Compose honours the services'
  dependencies, e.g. healthy databases and completed migrations. A failed attempt is retried after 10 s, 30 s,
  1 min, then every 5 min, up to `containers.start_attempts`. The core-manager stays up either way.
- **Crashed containers:** after a successful start, the long-running containers are checked every
  `containers.check_interval`. One that is stopped, dead or never started in two checks in a row is started
  again, at most `containers.restart_limit` times within `containers.restart_window`, then it is left alone until
  the core-manager restarts. One-shot services (label `core_srv_oneshot: "true"`) are left alone.
- **Stop:** when the core-manager stops, it ends the start or the monitor and runs
  `docker compose -f <compose_file_path> stop`. To stop the core, stop the core-manager: a container stopped by
  hand is started again.

## Configuration

Besides the existing settings, `util/config.go` has:

| Key | Default | Meaning |
|---|---|---|
| `state_files.version_path` | empty | the installed core version |
| `state_files.update_path` | empty | the record of the last update (`key=value`) |
| `state_files.check_path` | empty | the record of the last version check (`key=value`) |
| `state_files.lock_path` | empty | the installer's lock file |
| `containers.docker_path` | `docker` | the docker CLI with the compose plugin |
| `containers.wait_sockets` | `[]` | unix sockets the containers use, waited for before each start |
| `containers.start_timeout` | 10 min | per start attempt, including the wait for the sockets |
| `containers.stop_timeout` | 2 min | for `compose stop` |
| `containers.start_attempts` | 12 | failed start attempts before giving up |
| `containers.check_interval` | 5 s | how often crashed containers are looked for |
| `containers.restart_limit` | 5 | restarts of a container within `restart_window` before giving up on it |
| `containers.restart_window` | 10 min | |

Durations are in nanoseconds. An empty state file path is not read.

## Health

`GET /health` (see the swagger docs for the fields) reports the core services and, from the installer's state
files:

- `version`: the content of the version file
- `update`, `check`: the last update and version check. The files are read tolerantly: unknown keys are
  ignored, missing keys stay empty, `warning=` lines are collected. `null` if there is no record, or the file
  can't be read or is larger than 64 KiB.
- `lock`: who holds the installer's lock (`holder`, `pid`, `since`). Reported only while the lock is held: the
  core-manager tries a non-blocking shared lock and releases it at once. `/proc/locks` can't be used, because it
  hides a lock once the `flock` command that took it has exited.
