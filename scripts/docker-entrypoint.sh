#!/bin/sh
set -eu

run_migrations() {
  case "${AUTO_MIGRATE:-false}" in
    1|true|TRUE|True|yes|YES|Yes|on|ON|On)
      echo "Running database migrations..."
      /app/bin/migrate \
        --config "${MIGRATE_CONFIG:-config/config.prod.yaml}" \
        --path "${MIGRATE_PATH:-migrations}" \
        up
      ;;
    0|false|FALSE|False|no|NO|No|off|OFF|Off|"")
      ;;
    *)
      echo "Invalid AUTO_MIGRATE value: ${AUTO_MIGRATE}; expected true or false" >&2
      exit 2
      ;;
  esac
}

if [ "$#" -eq 0 ]; then
  set -- server --config config/config.prod.yaml
elif [ "${1#-}" != "$1" ]; then
  set -- server "$@"
fi

command="$1"
shift

case "$command" in
  server)
    run_migrations
    exec /app/bin/server "$@"
    ;;
  scheduler|worker|migrate|tool)
    exec "/app/bin/${command}" "$@"
    ;;
  *)
    echo "Unknown command: ${command}" >&2
    echo "Available commands: server, scheduler, worker, migrate, tool" >&2
    exit 2
    ;;
esac
