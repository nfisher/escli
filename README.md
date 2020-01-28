## Usage

The environment variable `ESHOST` must be set to your Elasticsearch HTTP URL.

```
export ESHOST="http://localhost:9200"
```

### List Indexes

List indexes currently prints the index names for the cluster.

```
./escli ls
```

### Search

Search currently defaults to printing the first 15 Json documents in the given
index.

```
./escli search ${INDEX}
```