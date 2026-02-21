FROM postgres:16-bookworm@sha256:dc1229cd4629538dfde35a9f808fdba5a86e82ca947ff658f251db6fed03dc45

# Install cron extension
RUN apt-get update && apt-get install -y postgresql-16-cron \
    && rm -rf /var/lib/apt/lists/*

# Tell Postgres to preload pg_cron
CMD ["postgres", "-c", "shared_preload_libraries=pg_cron",  "-c", "cron.database_name=jobs"]