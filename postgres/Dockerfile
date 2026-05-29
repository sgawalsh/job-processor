FROM postgres:18-bookworm@sha256:501c9112cb737119b90618d7c09ad4eaab243e4b370050eb280061284cd2ed63

# Install cron extension
RUN apt-get update && apt-get install -y postgresql-18-cron \
    && rm -rf /var/lib/apt/lists/*

# Tell Postgres to preload pg_cron
CMD ["postgres", "-c", "shared_preload_libraries=pg_cron",  "-c", "cron.database_name=jobs"]