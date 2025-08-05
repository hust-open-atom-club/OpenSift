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
    dist_impact       double precision default 0,
    scores            double precision default 0,
    org_count         integer          default 0,
    _name             varchar(255),
    _owner            varchar(255),
    _source           varchar(255),
    need_update       boolean,
    license           varchar(255),
    language          varchar(255),
    clone_valid       boolean          default false not null,
    lang_eco_pagerank double precision default 0,
    lang_eco_impact   double precision default 0,
    dist_pagerank     double precision default 0
);

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

create table if not exists github_links
(
    git_link text not null
        primary key
);

create table if not exists arch_relationships
(
    frompackage varchar(255) not null
        references arch_packages,
    topackage   varchar(255) not null,
    primary key (frompackage, topackage)
);

create table if not exists debian_relationships
(
    frompackage varchar(255) not null
        references debian_packages,
    topackage   varchar(255) not null,
    primary key (frompackage, topackage)
);

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

create table if not exists gentoo_relationships
(
    frompackage varchar(255) not null
        references gentoo_packages,
    topackage   varchar(255) not null,
    primary key (frompackage, topackage)
);

create table if not exists homebrew_relationships
(
    frompackage varchar(255) not null
        references homebrew_packages,
    topackage   varchar(255) not null,
    primary key (frompackage, topackage)
);

create table if not exists nix_relationships
(
    frompackage varchar(255) not null
        references nix_packages,
    topackage   varchar(255) not null,
    primary key (frompackage, topackage)
);

create table if not exists git_repositories
(
    git_link            varchar(255) not null
        primary key,
    industry            integer,
    domestic            boolean,
    industry_confidence boolean,
    domestic_confidence boolean
);

create table if not exists git_relationships
(
    fromgitlink varchar(255) not null,
    togitlink   varchar(255) not null,
    primary key (fromgitlink, togitlink)
);

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
    link_confidence real,
    default_install integer          default 0,
    pkg_scores      double precision default 0
);

create table if not exists deepin_relationships
(
    frompackage varchar(255) not null
        references deepin_packages,
    topackage   varchar(255) not null,
    constraint debian_relationships_copy1_pkey
        primary key (frompackage, topackage)
);

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

create table if not exists ubuntu_relationships
(
    frompackage varchar(255) not null
        references ubuntu_packages,
    topackage   varchar(255) not null,
    constraint deepin_relationships_copy1_pkey
        primary key (frompackage, topackage)
);

create table if not exists fedora_packages
(
    package         text not null
        constraint homebrew_packages_copy1_pkey1
            primary key,
    homepage        text,
    description     text,
    depends_count   bigint           default 1,
    git_link        text,
    page_rank       double precision default 0,
    version         text,
    link_confidence real
);

create table if not exists fedora_relationships
(
    frompackage varchar(255) not null
        references fedora_packages,
    topackage   varchar(255) not null,
    constraint homebrew_relationships_copy1_pkey
        primary key (frompackage, topackage)
);

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

create table if not exists centos_relationships
(
    frompackage varchar(255) not null
        references centos_packages,
    topackage   varchar(255) not null,
    constraint fedora_relationships_copy1_pkey
        primary key (frompackage, topackage)
);

create table if not exists git_mirror
(
    git_link varchar(255),
    mirror1  varchar(255),
    mirror2  varchar(255),
    mirror3  varchar(255),
    others   varchar(255)
);

create table if not exists alpine_packages
(
    package         text not null
        constraint fedora_packages_copy1_pkey1
            primary key,
    homepage        text,
    description     text,
    depends_count   bigint           default 1,
    git_link        text,
    page_rank       double precision default 0,
    version         text,
    link_confidence real
);

create table if not exists alpine_relationships
(
    frompackage varchar(255) not null
        references alpine_packages,
    topackage   varchar(255) not null,
    constraint fedora_relationships_copy1_pkey1
        primary key (frompackage, topackage)
);

create table if not exists aur_packages
(
    package         text not null
        constraint alpine_packages_copy1_pkey
            primary key,
    homepage        text,
    description     text,
    depends_count   bigint           default 1,
    git_link        text,
    page_rank       double precision default 0,
    version         text,
    link_confidence real
);

create table if not exists aur_relationships
(
    frompackage varchar(255) not null
        references aur_packages,
    topackage   varchar(255) not null,
    constraint alpine_relationships_copy1_pkey
        primary key (frompackage, topackage)
);

create table if not exists _migrations_history
(
    id      integer generated always as identity,
    name    varchar,
    time    timestamp,
    version varchar
);

create view draw_arch(frompackage, topackage, fromdepends, todepends) as
SELECT ar.frompackage,
       ar.topackage,
       ap1.depends_count AS fromdepends,
       ap2.depends_count AS todepends
FROM arch_relationships ar
         JOIN arch_packages ap1 ON ar.frompackage::text = ap1.package
         JOIN arch_packages ap2 ON ar.topackage::text = ap2.package;

create view draw_debian(frompackage, topackage, fromdepends, todepends) as
SELECT ar.frompackage,
       ar.topackage,
       dp1.depends_count AS fromdepends,
       dp2.depends_count AS todepends
FROM debian_relationships ar
         JOIN debian_packages dp1 ON ar.frompackage::text = dp1.package
         JOIN debian_packages dp2 ON ar.topackage::text = dp2.package;

create view draw_gentoo(frompackage, topackage, fromdepends, todepends) as
SELECT ar.frompackage,
       ar.topackage,
       gp1.depends_count AS fromdepends,
       gp2.depends_count AS todepends
FROM gentoo_relationships ar
         JOIN gentoo_packages gp1 ON ar.frompackage::text = gp1.package
         JOIN gentoo_packages gp2 ON ar.topackage::text = gp2.package;

create view draw_homebrew(frompackage, topackage, fromdepends, todepends) as
SELECT ar.frompackage,
       ar.topackage,
       hp1.depends_count AS fromdepends,
       hp2.depends_count AS todepends
FROM homebrew_relationships ar
         JOIN homebrew_packages hp1 ON ar.frompackage::text = hp1.package
         JOIN homebrew_packages hp2 ON ar.topackage::text = hp2.package;

create view draw_nix(frompackage, topackage, fromdepends, todepends) as
SELECT ar.frompackage,
       ar.topackage,
       np1.depends_count AS fromdepends,
       np2.depends_count AS todepends
FROM nix_relationships ar
         JOIN nix_packages np1 ON ar.frompackage::text = np1.package
         JOIN nix_packages np2 ON ar.topackage::text = np2.package;
         
CREATE TABLE IF NOT EXISTS openkylin_packages
(
    package         TEXT NOT NULL
        CONSTRAINT openkylin_packages_pkey PRIMARY KEY,
    version         TEXT,
    homepage        TEXT,
    description     TEXT,
    depends_count   BIGINT           DEFAULT 1,
    git_link        TEXT,
    page_rank       DOUBLE PRECISION DEFAULT 0,
    link_confidence REAL
);


CREATE TABLE IF NOT EXISTS openkylin_relationships
(
    frompackage VARCHAR(255) NOT NULL REFERENCES openkylin_packages,
    topackage   VARCHAR(255) NOT NULL,
    CONSTRAINT openkylin_relationships_pkey PRIMARY KEY (frompackage, topackage)
);


CREATE TABLE IF NOT EXISTS openeuler_packages
(
    package         TEXT NOT NULL
        CONSTRAINT openeuler_packages_pkey PRIMARY KEY,
    version         TEXT,
    homepage        TEXT,
    description     TEXT,
    depends_count   BIGINT           DEFAULT 1,
    git_link        TEXT,
    page_rank       DOUBLE PRECISION DEFAULT 0,
    link_confidence REAL
);


CREATE TABLE IF NOT EXISTS openeuler_relationships
(
    frompackage VARCHAR(255) NOT NULL REFERENCES openeuler_packages,
    topackage   VARCHAR(255) NOT NULL,
    CONSTRAINT openeuler_relationships_pkey PRIMARY KEY (frompackage, topackage)
);


CREATE TABLE IF NOT EXISTS openanolis_packages
(
    package         TEXT NOT NULL
        CONSTRAINT openanolis_packages_pkey PRIMARY KEY,
    version         TEXT,
    homepage        TEXT,
    description     TEXT,
    depends_count   BIGINT           DEFAULT 1,
    git_link        TEXT,
    page_rank       DOUBLE PRECISION DEFAULT 0,
    link_confidence REAL
);


CREATE TABLE IF NOT EXISTS openanolis_relationships
(
    frompackage VARCHAR(255) NOT NULL REFERENCES openanolis_packages,
    topackage   VARCHAR(255) NOT NULL,
    CONSTRAINT openanolis_relationships_pkey PRIMARY KEY (frompackage, topackage)
);


CREATE TABLE IF NOT EXISTS opencloudos_packages
(
    package         TEXT NOT NULL
        CONSTRAINT opencloudos_packages_pkey PRIMARY KEY,
    version         TEXT,
    homepage        TEXT,
    description     TEXT,
    depends_count   BIGINT           DEFAULT 1,
    git_link        TEXT,
    page_rank       DOUBLE PRECISION DEFAULT 0,
    link_confidence REAL
);


CREATE TABLE IF NOT EXISTS opencloudos_relationships
(
    frompackage VARCHAR(255) NOT NULL REFERENCES opencloudos_packages,
    topackage   VARCHAR(255) NOT NULL,
    CONSTRAINT opencloudos_relationships_pkey PRIMARY KEY (frompackage, topackage)
);


CREATE TABLE IF NOT EXISTS openharmony_packages
(
    package         TEXT NOT NULL
        CONSTRAINT openharmony_packages_pkey PRIMARY KEY,
    version         TEXT,
    homepage        TEXT,
    description     TEXT,
    depends_count   BIGINT           DEFAULT 1,
    git_link        TEXT,
    page_rank       DOUBLE PRECISION DEFAULT 0,
    link_confidence REAL
);


CREATE TABLE IF NOT EXISTS openharmony_relationships
(
    frompackage VARCHAR(255) NOT NULL REFERENCES openharmony_packages,
    topackage   VARCHAR(255) NOT NULL,
    CONSTRAINT openharmony_relationships_pkey PRIMARY KEY (frompackage, topackage)
);
         
         

