create table if not exists git_metrics_history
(
    id                       integer generated always as identity
        primary key,
    git_link                 varchar(255) not null,
    ecosystem                varchar[],
    created_since            date,
    updated_since            date,
    contributor_count        integer          default 0,
    commit_frequency         double precision default 0,
    depsdev_count            integer          default 0,
    deps_distro              double precision default 0,
    org_count                integer          default 0,
    _name                    varchar(255),
    _owner                   varchar(255),
    _source                  varchar(255),
    license                  varchar(255),
    language                 varchar[],
    clone_valid              boolean          default false,
    is_deleted               boolean          default false,
    depsdev_pagerank         double precision default 0,
    scores                   double precision default 0,
    update_time_git_metadata timestamp,
    update_time_deps_dev     timestamp,
    update_time_distribution timestamp,
    update_time_scores       timestamp,
    update_time              timestamp
);

create index if not exists idx_git_metrics_git_link
    on git_metrics_history (git_link);

insert into git_metrics_history
(git_link, ecosystem, created_since, updated_since, _name, _owner, _source,
 license, language, update_time_git_metadata, update_time_deps_dev, update_time_distribution, update_time_scores,
 update_time, scores)
select git_link,
       string_to_array(ecosystem, ' '),
       created_since,
       updated_since,
       _name,
       _owner,
       _source,
       license,
       string_to_array(language, ' '),
       now(),
       now(),
       now(),
       now(),
       now(),
       scores
from git_metrics;