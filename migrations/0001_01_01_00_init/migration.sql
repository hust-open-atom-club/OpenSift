create table if not exists arch_packages
(
    package         text not null
        primary key,
    version         text,
    depends_count   bigint           default 1,
    homepage        text,
    description     text,
    git_link        text,
    page_rank       double precision default 0,
    link_confidence real
);

alter table arch_packages
    owner to postgres;

create table if not exists debian_packages
(
    package         text not null
        primary key,
    version         text,
    homepage        text,
    description     text,
    depends_count   bigint           default 1,
    git_link        text,
    page_rank       double precision default 0,
    link_confidence real
);

alter table debian_packages
    owner to postgres;

create table if not exists git_metrics
(
    git_link          varchar(255)                   not null
        primary key,
    ecosystem         varchar(255),
    "from"            integer                        not null,
    created_since     date,
    updated_since     date,
    contributor_count integer          default 0,
    commit_frequency  double precision default 0,
    depsdev_count     integer          default 0,
    deps_distro       double precision default 0,
    scores            double precision default 0,
    org_count         integer          default 0,
    _name             varchar(255),
    _owner            varchar(255),
    _source           varchar(255),
    need_update       boolean,
    license           varchar(255),
    language          varchar(255),
    clone_valid       boolean          default false not null,
    depsdev_pagerank  double precision default 0
);

alter table git_metrics
    owner to postgres;

create table if not exists nix_packages
(
    package         text not null
        constraint arch_packages_copy1_pkey
            primary key,
    version         text,
    depends_count   bigint           default 1,
    homepage        text,
    description     text,
    git_link        text,
    page_rank       double precision default 0,
    link_confidence real
);

alter table nix_packages
    owner to postgres;

create table if not exists homebrew_packages
(
    package         text not null
        constraint debian_packages_copy1_pkey
            primary key,
    homepage        text,
    description     text,
    depends_count   bigint           default 1,
    git_link        text,
    page_rank       double precision default 0,
    link_confidence real
);

alter table homebrew_packages
    owner to postgres;

create table if not exists github_links
(
    git_link text not null
        primary key
);

alter table github_links
    owner to postgres;

create table if not exists arch_relationships
(
    frompackage varchar(255) not null
        references arch_packages,
    topackage   varchar(255) not null,
    primary key (frompackage, topackage)
);

alter table arch_relationships
    owner to postgres;

create table if not exists debian_relationships
(
    frompackage varchar(255) not null
        references debian_packages,
    topackage   varchar(255) not null,
    primary key (frompackage, topackage)
);

alter table debian_relationships
    owner to postgres;

create table if not exists gentoo_packages
(
    package         text not null
        constraint homebrew_packages_copy1_pkey
            primary key,
    version         text,
    homepage        text,
    description     text,
    depends_count   bigint           default 1,
    git_link        text,
    link_confidence real,
    page_rank       double precision default 0
);

alter table gentoo_packages
    owner to postgres;

create table if not exists gentoo_relationships
(
    frompackage varchar(255) not null
        references gentoo_packages,
    topackage   varchar(255) not null,
    primary key (frompackage, topackage)
);

alter table gentoo_relationships
    owner to postgres;

create table if not exists homebrew_relationships
(
    frompackage varchar(255) not null
        references homebrew_packages,
    topackage   varchar(255) not null,
    primary key (frompackage, topackage)
);

alter table homebrew_relationships
    owner to postgres;

create table if not exists nix_relationships
(
    frompackage varchar(255) not null
        references nix_packages,
    topackage   varchar(255) not null,
    primary key (frompackage, topackage)
);

alter table nix_relationships
    owner to postgres;

create table if not exists git_repositories
(
    git_link            varchar(255) not null
        primary key,
    industry            integer,
    domestic            boolean,
    industry_confidence boolean,
    domestic_confidence boolean
);

alter table git_repositories
    owner to postgres;

create table if not exists git_relationships
(
    fromgitlink varchar(255) not null,
    togitlink   varchar(255) not null,
    primary key (fromgitlink, togitlink)
);

alter table git_relationships
    owner to postgres;

create table if not exists git_metrics_prod
(
    git_link          varchar(255) not null
        constraint git_metrics_copy1_pkey
            primary key,
    ecosystem         varchar(255),
    "from"            integer      not null,
    created_since     date,
    updated_since     date,
    contributor_count integer          default 0,
    commit_frequency  double precision default 0,
    depsdev_count     integer          default 0,
    deps_distro       double precision default 0,
    scores            double precision default 0,
    org_count         integer          default 0,
    _name             varchar(255),
    _owner            varchar(255),
    _source           varchar(255),
    need_update       boolean,
    license           varchar(255),
    language          varchar(255)
);

alter table git_metrics_prod
    owner to postgres;

create table if not exists deepin_packages
(
    package         text not null
        constraint debian_packages_copy1_pkey1
            primary key,
    version         text,
    homepage        text,
    description     text,
    depends_count   bigint           default 1,
    git_link        text,
    page_rank       double precision default 0,
    link_confidence real
);

alter table deepin_packages
    owner to postgres;

create table if not exists deepin_relationships
(
    frompackage varchar(255) not null
        references deepin_packages,
    topackage   varchar(255) not null,
    constraint debian_relationships_copy1_pkey
        primary key (frompackage, topackage)
);

alter table deepin_relationships
    owner to postgres;

create table if not exists ubuntu_packages
(
    package         text not null
        constraint deepin_packages_copy1_pkey
            primary key,
    version         text,
    homepage        text,
    description     text,
    depends_count   bigint           default 1,
    git_link        text,
    page_rank       double precision default 0,
    link_confidence real
);

alter table ubuntu_packages
    owner to postgres;

create table if not exists ubuntu_relationships
(
    frompackage varchar(255) not null
        references ubuntu_packages,
    topackage   varchar(255) not null,
    constraint deepin_relationships_copy1_pkey
        primary key (frompackage, topackage)
);

alter table ubuntu_relationships
    owner to postgres;

create table if not exists fedora_packages
(
    package       text not null
        constraint homebrew_packages_copy1_pkey1
            primary key,
    homepage      text,
    description   text,
    depends_count bigint           default 1,
    git_link      text,
    page_rank     double precision default 0,
    version       text
);

alter table fedora_packages
    owner to postgres;

create table if not exists fedora_relationships
(
    frompackage varchar(255) not null
        references fedora_packages,
    topackage   varchar(255) not null,
    constraint homebrew_relationships_copy1_pkey
        primary key (frompackage, topackage)
);

alter table fedora_relationships
    owner to postgres;

create table if not exists centos_packages
(
    package         text not null
        constraint fedora_packages_copy1_pkey
            primary key,
    homepage        text,
    description     text,
    depends_count   bigint           default 1,
    git_link        text,
    page_rank       double precision default 0,
    version         text,
    link_confidence real
);

alter table centos_packages
    owner to postgres;

create table if not exists centos_relationships
(
    frompackage varchar(255) not null
        references centos_packages,
    topackage   varchar(255) not null,
    constraint fedora_relationships_copy1_pkey
        primary key (frompackage, topackage)
);

alter table centos_relationships
    owner to postgres;

create table if not exists git_mirror
(
    git_link varchar(255),
    mirror1  varchar(255),
    mirror2  varchar(255),
    mirror3  varchar(255),
    others   varchar(255)
);

alter table git_mirror
    owner to postgres;

create table if not exists alpine_packages
(
    package       text not null
        constraint fedora_packages_copy1_pkey1
            primary key,
    homepage      text,
    description   text,
    depends_count bigint           default 1,
    git_link      text,
    page_rank     double precision default 0,
    version       text
);

alter table alpine_packages
    owner to postgres;

create table if not exists alpine_relationships
(
    frompackage varchar(255) not null
        references alpine_packages,
    topackage   varchar(255) not null,
    constraint fedora_relationships_copy1_pkey1
        primary key (frompackage, topackage)
);

alter table alpine_relationships
    owner to postgres;

create table if not exists aur_packages
(
    package       text not null
        constraint alpine_packages_copy1_pkey
            primary key,
    homepage      text,
    description   text,
    depends_count bigint           default 1,
    git_link      text,
    page_rank     double precision default 0,
    version       text
);

alter table aur_packages
    owner to postgres;

create table if not exists aur_relationships
(
    frompackage varchar(255) not null
        references aur_packages,
    topackage   varchar(255) not null,
    constraint alpine_relationships_copy1_pkey
        primary key (frompackage, topackage)
);

alter table aur_relationships
    owner to postgres;

create table if not exists _migrations_history
(
    id      integer generated always as identity,
    name    varchar,
    time    timestamp,
    version varchar
);

alter table _migrations_history
    owner to postgres;
