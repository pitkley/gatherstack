# `src repos delete-metadata`


## Flags

| Name | Description | Default Value |
|------|-------------|---------------|
| `-dump-requests` | Log GraphQL requests and responses to stdout | `false` |
| `-get-curl` | Print the curl command for executing this query and exit (WARNING: includes printing your access token!) | `false` |
| `-insecure-skip-verify` | Skip validation of TLS certificates against trusted chains | `false` |
| `-key` | The name of the key to be deleted (required) |  |
| `-repo` | The ID of the repo with the key-value pair to be deleted (required) |  |
| `-trace` | Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing | `false` |
| `-user-agent-telemetry` | Include the operating system and architecture in the User-Agent sent with requests to Sourcegraph | `true` |


## Usage

```
Usage of 'src repos delete-metadata':
  -dump-requests
    	Log GraphQL requests and responses to stdout
  -get-curl
    	Print the curl command for executing this query and exit (WARNING: includes printing your access token!)
  -insecure-skip-verify
    	Skip validation of TLS certificates against trusted chains
  -key string
    	The name of the key to be deleted (required)
  -repo string
    	The ID of the repo with the key-value pair to be deleted (required)
  -trace
    	Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing
  -user-agent-telemetry
    	Include the operating system and architecture in the User-Agent sent with requests to Sourcegraph (default true)

Examples:

  Delete a key-value pair metadata from a repository:

    	$ src repos delete-metadata -repo=repoID -key=mykey



```
	