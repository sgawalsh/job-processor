#v16
FROM postgres@sha256:803ea6da631eab4a350da4ba87541e4737197b8c26ec0feb23ff6f9714212d1d

# Install cron extension
RUN apt-get update && apt-get install -y postgresql-16-cron \
    && rm -rf /var/lib/apt/lists/*

# Tell Postgres to preload pg_cron
CMD ["postgres", "-c", "shared_preload_libraries=pg_cron",  "-c", "cron.database_name=jobs"]