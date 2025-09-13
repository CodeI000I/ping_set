## About the project

ping_set is a Go utility for checking HTTP statuses of a list of URLs and displaying the result in a table for a quick overview of link availability.

Suitable for quick link inspection in local lists, reports, and CI pipelines, where a compact tabular output of HTTP statuses is important.
The project is implemented in Go, which simplifies cross-platform assembly and distribution of the binary.

### Features

- Checking HTTP statuses for a set of URLs with a tabular output of the result.

- A simple CLI utility, convenient for scripts and integration into CI.

- Cross-platform assembly using the standard go toolchain.

### Instalation

The utility can be installed with a single command with versioning via the @latest suffix:

```bash
go install github.com/CodeI000I/ping_set@latest
```

After installation, the binary will appear in the GOPATH/bin directory or GOBIN, if specified.

### Usage

It is recommended to store the list of input URLs in a text file, one URL per line.

Typical scenarios are shown below; if necessary, adjust the flags to suit the actual implementation of the program.

```bash
# Example 1: reading from file via flag (not implemented yet)
ping_set -f urls.txt

# Example 2: reading from STDIN (if implemented)
cat urls.txt | ping_set
```

The output is a tabular format with website domen, HTTP status and short description columns.
Example of input file

Each line is one URL.

```text
https://example.com
https://example.org/docs
http://localhost:8080/health
```

Example of expected output

```markdown
| SITE                | STATUS CODE  | DESCRIPTION                                             |
|---------------------|--------------|---------------------------------------------------------|
| facebook.com        | net_op_error | connect: connection refused                             |
| instagram.com       | net_op_error | connect: connection refused                             |
| medium.com          | net_op_error | read: connection reset by peer                          |
| stackexchange.com   | net_op_error | connect: connection refused                             |
| clevelandclinic.org | net_op_error | connect: connection refused                             |
| dailymotion.com     | dns_error    | lookup dailymotion.com on 192.168.1.1:53: no such host  |
| bbc.com             | net_op_error | read: connection reset by peer                          |
| indeed.com          | 403          | Forbidden                                               |
| scribd.com          | unknown code | unknown                                                 |
| youtube.com         | 200          | OK                                                      |
| healthline.com      | 403          | Forbidden                                               |
```
### Contributions

Suggestions, improvements, and fixes are welcome; create an Issue or open a Pull Request in the standard GitHub workflow.

Please avoid regressions and split changes into atomic commits to simplify review.
