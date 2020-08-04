# list-bq-permissions

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
