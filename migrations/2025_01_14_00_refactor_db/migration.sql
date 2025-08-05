-- ****** BEGIN migrate distribution_dependencies

create table distribution_dependencies
(
    id          int8 primary key generated always as identity,
    git_link    varchar,
    type        int,
    dep_count   int,
    impact      float8,
    page_rank   float8,
    update_time timestamp
);

create index on distribution_dependencies (git_link);

-- Debian 0
-- Arch 1
-- Homebrew 2
-- Nix 3
-- Alpine 4
-- Centos 5
-- Aur 6
-- Deepin 7
-- Fedora 8
-- Gentoo 9
-- Ubuntu 10
-- OpenEuler 11
-- OpenKylin 12
-- OpenAnolis 13
-- OpenCloudOS 14
-- OpenHarmony 15

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 0, depends_count, page_rank, 0, now()
from debian_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 1, depends_count, page_rank, 0, now()
from arch_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 2, depends_count, page_rank, 0, now()
from homebrew_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 3, depends_count, page_rank, 0, now()
from nix_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 4, depends_count, page_rank, 0, now()
from alpine_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 5, depends_count, page_rank, 0, now()
from centos_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 6, depends_count, page_rank, 0, now()
from aur_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 7, depends_count, page_rank, 0, now()
from deepin_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 8, depends_count, page_rank, 0, now()
from fedora_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 9, depends_count, page_rank, 0, now()
from gentoo_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 10, depends_count, page_rank, 0, now()
from ubuntu_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 11, depends_count, page_rank, 0, now()
from openeuler_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 12, depends_count, page_rank, 0, now()
from openkylin_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 13, depends_count, page_rank, 0, now()
from openanolis_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 14, depends_count, page_rank, 0, now()
from opencloudos_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

insert into distribution_dependencies (git_link, type, dep_count, page_rank, impact, update_time)
select git_link, 15, depends_count, page_rank, 0, now()
from openharmony_packages
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';


-- ****** END migrate distribution_dependencies


-- ****** BEGIN git metrics
create table git_metrics_tmp
(
    id                       int8 generated always as identity
        primary key,
    git_link                 varchar not null,
    created_since            date,
    updated_since            date,
    contributor_count        integer          default 0,
    commit_frequency         double precision default 0,
    org_count                integer          default 0,
    license                  varchar[],
    language                 varchar[],
    clone_valid              boolean          default false,
    update_time              timestamp
);

create index if not exists idx_git_metrics_git_link
    on git_metrics_tmp (git_link);

insert into git_metrics_tmp (git_link, created_since, updated_since, contributor_count, commit_frequency, org_count, license, language, clone_valid, update_time)
select git_link, created_since, updated_since, contributor_count, commit_frequency, org_count, string_to_array(license, ' '), string_to_array(language, ' '), clone_valid, now() from git_metrics;

drop table git_metrics;

alter table git_metrics_tmp rename to git_metrics;
alter table git_metrics rename constraint git_metrics_tmp_pkey to git_metrics_pkey;

-- ****** END git metrics

-- ****** BEGIN git mirror set
create table git_mirror_set
(
    id          int8 primary key generated always as identity,
    git_link    varchar not null unique check (git_link <> ''),
    parent      varchar not null
);
-- ****** END git mirror set

-- ****** BEGIN lang_ecosystems

-- npm 0
-- go 1
-- maven 2
-- pypi 3
-- nuget 4
-- cargo 5

create table lang_ecosystems
(
    id          int8 primary key generated always as identity,
    git_link    varchar not null,
    type        int,
    dep_count   int,
    lang_eco_impact float8,
    update_time timestamp default now()
);

-- ****** END lang_ecosystems

-- ****** BEGIN platform_links
create table gitlab_links
(
    git_link varchar not null unique check (git_link <> '') primary key
);

create table bitbucket_links
(
    git_link varchar not null unique check (git_link <> '') primary key
);
-- ****** END platform_links


-- ****** BEGIN create all_gitlinks view
create index on debian_packages (git_link);
create index on arch_packages (git_link);
create index on homebrew_packages (git_link);
create index on nix_packages (git_link);
create index on alpine_packages (git_link);
create index on centos_packages (git_link);
create index on aur_packages (git_link);
create index on deepin_packages (git_link);
create index on fedora_packages (git_link);
create index on gentoo_packages (git_link);
create index on ubuntu_packages (git_link);
create index on openeuler_packages (git_link);
create index on openkylin_packages (git_link);
create index on openanolis_packages (git_link);
create index on opencloudos_packages (git_link);
create index on openharmony_packages (git_link);

create view all_gitlinks as
select git_link from (
                         select distinct git_link from debian_packages
                         union distinct select git_link from arch_packages
                         union distinct select git_link from homebrew_packages
                         union distinct select git_link from nix_packages
                         union distinct select git_link from alpine_packages
                         union distinct select git_link from centos_packages
                         union distinct select git_link from aur_packages
                         union distinct select git_link from deepin_packages
                         union distinct select git_link from fedora_packages
                         union distinct select git_link from gentoo_packages
                         union distinct select git_link from ubuntu_packages
                         union distinct select git_link from openeuler_packages
                         union distinct select git_link from openkylin_packages
                         union distinct select git_link from openanolis_packages
                         union distinct select git_link from opencloudos_packages
                         union distinct select git_link from openharmony_packages
                         union distinct select git_link from github_links
                         union distinct select git_link from gitlab_links
                         union distinct select git_link from bitbucket_links) t
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';

-- ****** END create all_gitlinks view

-- ****** BEGIN score

create table scores
(
    id          int8 primary key generated always as identity,
    git_link    varchar not null,
    dist_score  float8,
    lang_score  float8,
    git_score   float8,
    score       float8,
    update_time timestamp default now()
);

create table scores_dist
(
    score_id   int8 generated always as identity references scores (id),
    distribution_dependencies_id int8 references distribution_dependencies (id),

    primary key (score_id, distribution_dependencies_id)
);

create table scores_lang
(
    score_id   int8 generated always as identity references scores (id),
    lang_ecosystems_id int8 references lang_ecosystems (id),

    primary key (score_id, lang_ecosystems_id)
);

create table scores_git (
    score_id   int8 generated always as identity references scores (id),
    git_metrics_id int8 references git_metrics (id),

    primary key (score_id, git_metrics_id)
);

-- ****** END score

-- ****** BEGIN workflows
create table workflows
(
    id          int8 primary key generated always as identity,
    job_id      varchar not null,
    task_name   varchar not null,
    action      varchar not null,
    payload     jsonb,
    update_time timestamp default now()
);

-- drop useless tables
drop table git_metrics_prod;

create view result as
select s.id as score_id,
       min(s.git_link) as git_link,
       array_agg(gm.language) as language,
       array_agg(gm.license) as license,
       -- git
       max(gm.created_since) as git_created_since,
       max(gm.updated_since) as git_updated_since,
       max(gm.contributor_count) as git_contributor_count,
       max(gm.commit_frequency) as git_commit_frequency,
       max(gm.org_count) as git_org_count,
       max(s.git_score) as git_score,
       -- dist
       sum(dd.dep_count) as dist_count,
       sum(dd.impact) as dist_impact,
       sum(dd.page_rank) as dist_page_rank,
       max(s.dist_score) as dist_score,
       -- lang
       sum(le.dep_count) as lang_count,
       max(s.lang_score) as lang_score,
       -- score
       max(s.score) as score,
       max(s.update_time) as update_time
from (
    select distinct on (git_link) *
    from scores
    order by git_link, id desc
) s left join scores_git sg on s.id = sg.score_id
    left join git_metrics gm on gm.id = sg.git_metrics_id
    left join scores_dist sd on s.id = sd.score_id
    left join distribution_dependencies dd on sd.distribution_dependencies_id = dd.id
    left join scores_lang sl on s.id = sl.score_id
    left join lang_ecosystems le on sl.lang_ecosystems_id = le.id
group by s.id
order by max(s.score) desc;


create function result_until(until timestamp)
    returns table (
        score_id           int8,
        git_link           varchar,
        language           varchar[],
        license            varchar[],
        git_created_since  timestamp,
        git_updated_since  timestamp,
        git_contributor_count int8,
        git_commit_frequency float8,
        git_org_count      int8,
        git_score          float8,
        dist_count         int8,
        dist_impact        float8,
        dist_page_rank     float8,
        dist_score         float8,
        lang_count         int8,
        lang_score         float8,
        score              float8,
        update_time        timestamp
    ) as
$body$
select s.id as score_id,
       min(s.git_link) as git_link,
       array_agg(gm.language) as language,
       array_agg(gm.license) as license,
       -- git
       max(gm.created_since) as git_created_since,
       max(gm.updated_since) as git_updated_since,
       max(gm.contributor_count) as git_contributor_count,
       max(gm.commit_frequency) as git_commit_frequency,
       max(gm.org_count) as git_org_count,
       max(s.git_score) as git_score,
       -- dist
       sum(dd.dep_count) as dist_count,
       sum(dd.impact) as dist_impact,
       sum(dd.page_rank) as dist_page_rank,
       max(s.dist_score) as dist_score,
       -- lang
       sum(le.dep_count) as lang_count,
       max(s.lang_score) as lang_score,
       -- score
       max(s.score) as score,
       max(s.update_time) as update_time
from (
         select distinct on (git_link) *
         from scores
         where update_time <= until
         order by git_link, id desc
     ) s left join scores_git sg on s.id = sg.score_id
         left join git_metrics gm on gm.id = sg.git_metrics_id
         left join scores_dist sd on s.id = sd.score_id
         left join distribution_dependencies dd on sd.distribution_dependencies_id = dd.id
         left join scores_lang sl on s.id = sl.score_id
         left join lang_ecosystems le on sl.lang_ecosystems_id = le.id
group by s.id
$body$
language sql;
