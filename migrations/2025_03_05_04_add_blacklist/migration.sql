create table if not exists git_link_blacklist (
    git_link text not null primary key
);

create or replace view all_gitlinks as
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
                         union distinct select git_link from bitbucket_links
                         except select git_link from git_link_blacklist) t
where git_link is not null and git_link <> '' and git_link <> 'NA' and git_link <> 'NaN';
