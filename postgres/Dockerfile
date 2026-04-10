FROM postgres:18-bookworm@sha256:7c7922bce7110e690e5a48d241e7fedd0daf6b4622dddf2f41f659a71ee445c4

# Install cron extension
RUN apt-get update && apt-get install -y postgresql-18-cron \
    && rm -rf /var/lib/apt/lists/*

# Tell Postgres to preload pg_cron
CMD ["postgres", "-c", "shared_preload_libraries=pg_cron",  "-c", "cron.database_name=jobs"]