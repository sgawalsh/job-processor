FROM postgres:18-bookworm@sha256:db87f88e316d16637f9f906ed83c0118f81ef1cecb1680b77bb0f4b7187643d5

# Install cron extension
RUN apt-get update && apt-get install -y postgresql-16-cron \
    && rm -rf /var/lib/apt/lists/*

# Tell Postgres to preload pg_cron
CMD ["postgres", "-c", "shared_preload_libraries=pg_cron",  "-c", "cron.database_name=jobs"]