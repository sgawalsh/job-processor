FROM postgres:18-bookworm@sha256:a8824b3eef0c73d7a494881d625411060f5bf3c87280394cbc84197c6fa4bde5

# Install cron extension
RUN apt-get update && apt-get install -y postgresql-18-cron \
    && rm -rf /var/lib/apt/lists/*

# Tell Postgres to preload pg_cron
CMD ["postgres", "-c", "shared_preload_libraries=pg_cron",  "-c", "cron.database_name=jobs"]