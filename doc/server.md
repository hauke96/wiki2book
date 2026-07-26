> This feature is **EXPERIMENTAL**.

Wiki2book can be started as an HTTP server offering a similar API to the CLI.

Currently, there is no version schema, no guarantee for backward compatibility and not everything can be set/configured using the HTTP API.
These things might change over time (see the [section about issues](#known-and-open-issues) below).

The general thought is to provide a simple API, which might grow and evolve over time.
This also means that wiki2book does not and will not handle SSL, load balancing or further security measurements.
Use appropriate tools for that (e.g. nginx with appropriate configuration for HTTPS and zoning for DoS protection).

Normal wiki2book configurations (e.g. caching or command templates) are picked up and used just like the CLI does.
See the [config section](#configuration) for details.
This means it should be fairly safe (no guarantee!) to host wiki2book online.
But keep in mind: The more people use such an online server, the faster will the Wikipedia API rate limit kick in.
Use this feature responsibly!

# API

## General idea

Because the creation of eBooks might take a while, the creation is asynchronous.
Every of the endpoints to create an eBook will quite immediately return a response with a

## Endpoints

### `GET /article/{articleName}`

Creates an eBook of the given article.
The server uses its default configuration.

#### Parameter

* `articleName`: The name of the Wikipedia article.

#### Response

* In case of success: A 200 response with a JSON `ResultState` object (s. [data section](#resultstate)) as body.
* In case of an error: A simple text describing the error.

### `POST /article/{articleName}`

Creates an eBook of the given article.
The server uses its default configuration.

#### Parameter

* `articleName`: The name of the Wikipedia article.

#### Request body

The content of a configuration file as `application/json` content.

**Note:** Some fields will not be picked up and will be reset to the servers defaults.
See [data section](#configuration-and-project-content) for details.

#### Response

* In case of success: A 200 response with a JSON `ResultState` object (s. [data section](#data)) as body.
* In case of an error: A simple text describing the error.

### `POST /project`

Creates an eBook based on a project file sent as request body.

#### Request body

The content of a project file as `application/json` content.

**Note:** Some fields will not be picked up and will be reset to the servers defaults.
See [data section](#configuration-and-project-content) for details.

#### Response

* In case of success: A 200 response with a JSON `ResultState` object (s. [data section](#data)) as body.
* In case of an error: A simple text describing the error.

### `POST /standalone`

Creates an eBook based on a form-data request containing wikitext and optionally a configuration file.

#### Request body

A `text/plain` or `multipart/form-data` body.
The `text/plain` body is interpreted as wikitext and the configuration of the server is used.
The `multipart/form-data` body must contain the two sections `config` and `content` for the JSON configuration and wikitext content.

**Note:** Some fields of the configuration will not be picked up and will be reset to the servers defaults.
See [data section](#configuration-and-project-content) for details.

#### Response

* In case of success: A 200 response with a JSON `ResultState` object (s. [data section](#data)) as body.
* In case of an error: A simple text describing the error.

### `GET /states/{resultToken}`

Gets the current [result state](#resultstate) object of the given token.

#### Parameter

* `resultToken`: The result state token returned when creating an eBook.

#### Response

* In case of success: A 200 response with a JSON `ResultState` object (s. [data section](#resultstate)) as body.
* In case of an error: A simple text describing the error.

### `GET /results/{resultToken}`

Gets the actual result of the eBook creation.
Will return an error when the result state is not `SUCCESS`.

#### Response

* In case of success: A 200 response with an `Content-Type: application/octet-stream` body containing the eBook file bytes.
  The `Content-Disposition` header contains the filename, which might just contain the token string.
  Example: `attachment;filename="b6e0c56635b2cebf1374d6c6e22770a322f53847"`.
* In case of an error: A simple text describing the error.

## Data

### `ResultState`

A JSON response with the following fields:

* `status`: The status of the eBook creation. One of `IN_PROGRESS`, `SUCCESS` or `FAILED`
* `title`: The title of the book.
* `result-token`: A random string that should be used as `resultToken` parameter to get an updated result state or the actual result.

#### Example

```json
{
  "status": "IN_PROGRESS",
  "title": "My eBook",
  "result-token": "b6e0c56635b2cebf1374d6c6e22770a322f53847"
}
```

### Configuration and project content

Some endpoints allow the upload of configuration or project files.
For security reasons, not all fields are picked up.
These fields are ignored and the servers default values are used:

* `allowed-link-prefixes`
* `category-prefixes`
* `file-prefixes`
* `ignored-image-params`
* `ignored-media-types`
* `ignored-templates`
* `output-type`
* `svg-size-to-viewbox`
* `toc-depth`
* `trailing-templates`
* `wikipedia-host`
* `wikipedia-image-article-hosts`
* `wikipedia-image-host`
* `wikipedia-instance`
* `wikipedia-math-rest-api`

See the [configuration.md](./configuration.md) for all possible values.

# Configuration

The configuration of the server is the same as for the CLI.
See the [configuration.md](./configuration.md) for details.
All server-specific config entries have the prefix "Server".

# Known and open issues

These are known and open issues, problems or missing features.
They will eventually be addressed over time.

* Not everything of a request is validated, e.g. the `articleName` parameter or the values of headers.
* The code still uses a lot of fatal error checks.
  Theses are error checks that would crash the application.
  For a CLI that's fine but a server should keep on running and, therefore, this is an open issue and needs refinement in the handling of errors.
* Uploading CSS styles, fonts, cover images etc. is not yet possible.
* The result state currently doesn't contain any error messages.