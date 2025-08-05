-- 逐个表修改字段名
ALTER TABLE debian_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE arch_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE homebrew_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE gentoo_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE alpine_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE nix_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE ubuntu_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE fedora_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE deepin_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE centos_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE aur_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE openeuler_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE openkylin_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE openanolis_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE opencloudos_packages RENAME COLUMN dl_3m_vol TO downloads_3m;
ALTER TABLE openharmony_packages RENAME COLUMN dl_3m_vol TO downloads_3m;