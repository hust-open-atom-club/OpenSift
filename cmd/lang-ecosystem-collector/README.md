# DepsDev Synchronization

## Installation

1. Clone this repository:

   ```
   git clone https://github.com/HUSTSecLab/OpenSift.git
   cd criticality_score
   ```

2. Install dependencies:

   - Ensure you have Go installed. You can install Go from [here](https://golang.org/doc/install).
   - Run `go mod tidy` to install required Go dependencies.

   ```
   go mod tidy
   ```

3. Set up PostgreSQL database:

   - Ensure you have PostgreSQL installed and running.
   - Configure the database with the appropriate tables (`git_metrics` with at least the `git_link` and `ecosystem` columns).
   - The application assumes that the database has a table called `git_metrics` that contains project information (e.g., GitHub, GitLab, Bitbucket links).

## Usage

1. **Run the synchronization script:**

   To start the synchronization process, use the following command:

   ```
   go run main.go --config=config.json
   ```

   This will:

   - Initialize the database connection using the provided `config.json`.
   - Fetch the Git project links from the `git_metrics` table.
   - For each project, query the [deps.dev API](https://api.deps.dev/) to retrieve the latest version and dependent information.
   - Update the `depsdev_count` in the `git_metrics` table with the dependent count for each project.