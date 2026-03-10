FROM postgres:18-bookworm@sha256:e83514f5fe85fab1692e05a510f3a10bc86a4ff428a333942239a72f9e0a5ce8

# Install cron extension
RUN apt-get update && apt-get install -y postgresql-18-cron \
    && rm -rf /var/lib/apt/lists/*

# Tell Postgres to preload pg_cron
CMD ["postgres", "-c", "shared_preload_libraries=pg_cron",  "-c", "cron.database_name=jobs"]