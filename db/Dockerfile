FROM postgres:16

# Install cron extension
RUN apt-get update && apt-get install -y postgresql-16-cron \
    && rm -rf /var/lib/apt/lists/*

# Tell Postgres to preload pg_cron
CMD ["postgres", "-c", "shared_preload_libraries=pg_cron",  "-c", "cron.database_name=jobs"]