# bqiam

[![Actions Status: golangci-lint](https://github.com/hirosassa/bqiam/workflows/golangci-lint/badge.svg)](https://github.com/hirosassa/bqiam/actions?query=workflow%3A"golangci-lint")
[![Apache-2.0](https://img.shields.io/github/license/hirosassa/bqiam)](LICENSE)


## What is this?

This tool provides easier permission management for BigQuery.

Currently supports;

- list the user's permissions for each BigQuery Datasets
- permit users to each BigQuery Datasets access role (READER/WRITER/OWNER) and `roles/bigquery.jobUser` (to run query)
- permit users to Project-wide access role (`roles/viewer` or `rolse/editor`)
- revoke users' datasets and/or project-wide permissions

## Requirement

You must have a `roles/owner` on your GCP project.


## Install

```bash
$ go install github.com/hirosassa/bqiam@latest
```


## Usage

Prepare configuration file as following format (currently support only the file name is `.bqiam.toml` on your `$HOME`):

```
// .bqiam.toml
BigqueryProjects = ["project-id-A", "project-id-B", ...]
CacheFile = "path/to/cache-file.toml"
```

Next, fetch bigquery dataset metadata and store it to cache file (take about 30-60 sec.).

```bash
$ bqiam cache
dataset meta data are cached to path/to/cache-file.toml
```

List datasets the user is able to access.
```bash
$ bqiam dataset abc@sample.com
sample-prj sample-ds1 OWNER
sample-prj sample-ds2 READER
...
```

Grant the user(s) a role to access the dataset(s). This command also adds `roles/bigquery.jobUser` automatically.

```bash
$ bqiam permit dataset READER -p bq-project-id -u user1@email.com -u user2@email.com -d dataset1 -d dataset2
Permit user1@email.com to dataset1 access as READER
Permit user2@email.com to dataset1 access as READER
...

```

Grant the user(s) a project-wide role.
```bash
$ bqiam permit project READER -p bq-project-id -u user1@email.com -u user2@email.com
Permit user1@email.com to bq-project-id access as READER
Permit user2@email.com to bq-project-id access as READER
...

```

Revoke the user(s)' access permissions.
```bash
$ bqiam revoke dataset READER -p bq-project-id -u user1@email.com -d dataset1
Revoked user1@email.com's permission of dataset1 access as READER

$ bqiam project READER -p bq-project-id -u user1@email.com -u user2@email.com
Revoked user1@email.com's permission of bq-project-id access as READER
Revoked user2@email.com's permission of bq-project-id access as READER
```


## Completion
Completion is available for bash or zsh.
Download projects, datasets, users list data via GCP API.

### Set up completion configurations
#### bash
```
bqiam completion bash > /path/to/bash-completion/completions/bqiam
```

#### zsh
```
bqiam completion zsh > /path/to/zsh-completions/_bqiam
```

### Prefetch / Updating completion data
Completion candidate data is downloaded to `~/.bqiam-completion-file.toml` by following command.
```
bqiam completion
```
