drop table if exists failed_git_metrics;

create table if not exists git_files (
    git_link text not null primary key,
    file_path text not null,
    success boolean not null,
    message text,
    update_time timestamptz,
    last_success timestamptz,
    take_time_ms int8,
    take_storage int8,
    failed_times int
)