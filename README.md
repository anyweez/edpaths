# ED Paths
Route planning tool for Elite Dangerous. Site coming soon.

## Tools

`go get github.com/anyweez/edpaths/...`

Two binaries are included in this repository:

- `spaceimp`, which imports system, station, and body data from eddb.io and forms a pair of local key-value stores that spacecrawl uses to plot routes.
- `spacecrawl`, which performs realtime queries based on user route specifications. Exposes a simple set of REST endpoints. Also offers system autocomplete.

Note that these tools do not fetch system / body / station data, but expect it to be availbable locally. You can download it yourself from [eddb's generous API page](https://eddb.io/api).