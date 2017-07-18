# The Reading List CLI

Use `gcloud auth application-default login` to generate credentials.

Alternatively, you can use the `-creds` flag that points to the path of a Google service account JSON key file.

## Usage

```
Usage of cli:
  -creds string
      the path of the service account credentials file. if empty, uses Google Application Default Credentials.
  -delete
      delete this URL from the list (requires -mode update)
  -host string
      the host of the reading list server (default "http://localhost:8080")
  -limit int
      limit for the number of links to return when listing links (default 20)
  -mode string
      (list|update) (default "list")
  -url string
      the URL to add or delete
```
