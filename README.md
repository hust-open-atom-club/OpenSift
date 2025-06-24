# OpenSift

An intelligent filtering engine to mine the most critical open source projects from massive repositories.

[简体中文](./README.zh_CN.md) 

## Description

This project aims to evaluate and rank open-source projects based on their criticality within the open source ecosystem and download volume. Unlike traditional methods that rely solely on GitHub metrics, this project incorporates data from various Linux distributions, corresponding package managers and more code hosting platforms to provide a more comprehensive analysis.

By collecting and analyzing metrics from multiple sources, this project offers a robust and comprehensive framework for assessing the criticality of open-source projects.

## Difference from [ossf/criticality_score](https://github.com/ossf/criticality_score)

- **Distribution Dependents**: Collects data from various Linux distributions (e.g. Debian, Arch, Nix, Gentoo) and corresponding package managers to evaluate the dependency of open-source software.
- **Support for All Git Repositories**: Analyzes repositories from any Git platform, not just GitHub.
- **Comprehensive Metrics Collection**: Gathers a wider and more precise metrics from Git repositories and package managers, for example, the number of commits, organization count is more accurate than GitHub API.
- **Friendly for Metrics Customization**: Customizes any metrics used in the criticality evaluation algorithm other than metrics that can be only collected by Github API.
- **No Dependency on Google Cloud or BigQuery**: `ossf/criticality_score` depends on Google Cloud service, making it hard to migrate to other platforms. This project runs independently of specific cloud services, ensuring ease of deployment.
- **Easy Deployment**: Runs a script, and the system will be easily setup with Docker.
- **Provides Additional Information**: Provides extra insights, such as relationships between projects and dependencies.

## Quick Start

If you want to collect data from Gentoo, please go to setup with Gentoo prefix, and refer to [How to setup Gentoo prefix](./docs/setup/gentoo.md).

Then make sure `docker` and `docker-compose-v2` is installed, and run the following commands:

```sh
export GENTOO_PREFIX_DIR=<your Gentoo prefix location> # If you don't have Gentoo prefix set, ignore
export GITHUB_TOKEN=<your GitHub token> # This is essential for github enumeration
./setup.sh
```

1. After finishing the setup script, try to connect to the postgresql database (the password is stored in `data/DB_PASSWD`).

2. Populate git_link fields in arch_packages, debian_packages and other distribution package table manually and finally run following command. If git_link data is already there, you can use `scripts/copy-gitlink.py` tool to copy the data to the database.

3. Execute the following command for the first time to collect and calculate the criticality score. This will take days to finish.

```sh
docker compose exec app bash /gitlink.sh
```

## Documentation of general design, tools and components

See [docs/](./docs/) for details

## Public Data

If you're interested in seeing a list of critical projects with their criticality
score, we publish them in `csv` format.

### CSV data

The data is available on HUST mirror site and can be downloaded via [this link](https://mirrors.hust.edu.cn/core-oss/criticality_score_data/), which includes:

- **[git_metrics_prod](https://mirrors.hust.edu.cn/core-oss/criticality_score_data/latest/git_metrics_prod.csv)**: The table contains important metrics and scores that are used to determine the criticality of various repositories. This table can help you assessment the criticality of repositories.
- **[git_relationships](https://mirrors.hust.edu.cn/core-oss/criticality_score_data/latest/git_relationships.csv)**: The table contains information about dependencies between Git repositories, forming a graph that represents how different repositories are related to each other. This table is essential for tracking the relationships and dependencies between various repositories.

## Metrics Description

We have established a model that includes three dimensions:

1. **Git Metadata**:
    | **Metric**           | **Description**                                                             | **Reasoning**                                                                          | **Threshold**  | **Weight**   |
    |----------------------|-----------------------------------------------------------------------------|---------------------------------------------------------------------------------------|----------------|--------------|
    | created_since        | The longer a project has existed, the more likely it is widely used or relied upon, representing maturity and stability. | Older projects tend to have a larger user base and more robust testing over time.     | 120 months     | 1           |
    | updated_since        | Projects not updated for a long time may no longer be maintained, reducing reliability and dependence. | Unmaintained projects are less likely to be secure or relevant for active use cases.  | 120 months     | -1            |
    | contributor_count    | A higher number of contributors indicates greater attention, community activity, and importance. | Diverse contributions demonstrate active engagement and widespread support.            | 40,000         | 2           |
    | org_count            | Contributions from multiple organizations indicate cross-organization dependencies and wide-ranging influence. | Cross-organizational contributions highlight the project’s universal relevance.        | 8,400          | 1         |
    | commit_frequency     | Higher code change frequency shows project activity but may also indicate potential vulnerability risks. | Active commits suggest responsiveness but may require monitoring for quality issues.   | 1,000 commits  | 1          |

2. **Distribution Dependencies**:
    | **Metric**           | **Description**                                                             | **Reasoning**                                                                          | **Threshold**  | **Weight**   |
    |----------------------|-----------------------------------------------------------------------------|---------------------------------------------------------------------------------------|----------------|--------------|
    | dist_impact          | Indicates how widely the project is relied upon in different software distributions, showing its production use and stability. | Broad use in distributions highlights the project’s practical utility and reliability. | 1              | 5          |
    | pagerank             | Calculated by the proportion of dependencies in each distribution and the corresponding package's PageRank score. | Higher PageRank in distributions indicates greater importance and influence.           | 1              | 5          |
    | default_install      | Indicates whether the project is included by default in the installation of some distributions (1 if included, 0 if not). **TODO**: Further details on the default installation metric will be provided. | Default inclusion in installations signifies the project's essential role and reliability. | 1              | 2.5        |

3. **Language Ecosystem (npm, pypi)**:
    | **Metric**           | **Description**                                                             | **Reasoning**                                                                          | **Threshold**  | **Weight**   |
    |----------------------|-----------------------------------------------------------------------------|---------------------------------------------------------------------------------------|----------------|--------------|
    | dist_impact          | Calculated by the proportion of dependencies in the language ecosystem, showing its importance in the development ecosystem. | Projects with more dependencies are critical to the development ecosystem.            | 1            | 5          |
    | pagerank             | Importance of each package in the dependency graph; we use the PageRank algorithm to calculate this metric. **TODO**: Further details on the PageRank calculation will be provided. | Projects with higher PageRank are more critical in the ecosystem.                     | 1            | 5        |


## Reference

[1] <https://github.com/ossf/criticality_score>
