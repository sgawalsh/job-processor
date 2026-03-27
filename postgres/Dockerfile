FROM postgres:18-bookworm@sha256:a73ac5911d8b2c3176bc10f8f9a963be7d825b9bbd8e55e652104b736c4db6e7

# Install cron extension
RUN apt-get update && apt-get install -y postgresql-18-cron \
    && rm -rf /var/lib/apt/lists/*

# Tell Postgres to preload pg_cron
CMD ["postgres", "-c", "shared_preload_libraries=pg_cron",  "-c", "cron.database_name=jobs"]