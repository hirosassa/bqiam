# bqiam

[![Actions Status: golangci-lint](https://github.com/hirosassa/bqiam/workflows/golangci-lint/badge.svg)](https://github.com/hirosassa/bqiam/actions?query=workflow%3A"golangci-lint")
[![Apache-2.0](https://img.shields.io/github/license/hirosassa/bqiam)](LICENSE)

## Usage

prepare configuration file as following format (currently support only the file name is `.bqiam.toml` on your `$HOME`):

```
BigqueryProjects = ["project-id-A", "project-id-B", ...]
CacheFile = "path/to/cache-file.toml"
```

then, you can use `bqiam` as follows:

```
$ bqiam cache  // fetch bigquery dataset metadata and store it to cache file (take about 30-60 sec.)
dataset meta data are cached to path/to/cache-file.toml

$ bqiam dataset "abc@sample.com"
sample-prj sample-ds1 OWNER
sample-prj sample-ds2 READER
...
```
