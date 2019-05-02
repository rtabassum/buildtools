# Buildifier

buildifier is a tool for formatting bazel BUILD and .bzl files with a standard convention.

## Setup

Build the tool:
* Checkout the repo and then either via `go install` or `bazel build //buildifier`
* If you already have 'go' installed, then build a binary via:

`go get github.com/bazelbuild/buildtools/buildifier`

## Usage

Use buildifier to create standardized formatting for BUILD and .bzl files in the
same way that clang-format is used for source files.

    $ buildifier path/to/file

You can also process multiple files at once:

    $ buildifier path/to/file1 path/to/file2

You can make buildifier automatically find all Starlark files (i.e. BUILD, WORKSPACE, .bzl, or .sky)
in a directory recursively:

    $ buildifier -r path/to/dir

Buildifier automatically detects the file type (either BUILD or .bzl) by its filename. If you 

    $ buildifier $(find . -type f \( -iname BUILD -or -iname BUILD.bazel \))

Files with unknown names (e.g. `foo.bar`) will be formatted as .bzl files because the format for
.bzl files is more flexible and less harmful.

You can use Buildifier as a filter by invoking it with no arguments. In that mode it reads from
standard input and writes the reformatted version to standard output. In this case it won't be
able to see its name to choose the correct formatting rules, and for compatibility reasons it
will use the BUILD format in such situations. This may be changed in the future, and to enforce
a special format explicitly use the `--type` flag:

    $ cat foo.bar | buildifier --type=build
    $ cat foo.baz | buildifier --type=bzl

## Linter

Buildifier has an integrated linter that can point out and in some cases automatically fix various
issues. To use it launch one of the following commands to show and to fix the issues
correspondingly:

    buildifier --lint=warn path/to/file
    buildifier --lint=fix path/to/file

By default the linter searches for all known issues except the following:

  * [out-of-order-load](../WARNINGS.md#out-of-order-load)
  * [unsorted-dict-items](../WARNINGS.md#unsorted-dict-items)

You can specify the categories using the `--warnings` flag either by providing the categories
explicitly:

    buildifier --lint=warn --warnings=positional-args,duplicated-name

or by modifying the default warnings set:

    buildifier --lint=warn --warnings=-positional-args,+unsorted-dict-items

It's also possible to provide `--warnings=all` to use all supported warnings categories.

See also the [full list](../WARNINGS.md) or the supported warnings.

## Setup and usage via Bazel

You can also invoke buildifier via the Bazel rule.
`WORKSPACE` file:
```bzl
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# buildifier is written in Go and hence needs rules_go to be built.
# See https://github.com/bazelbuild/rules_go for the up to date setup instructions.
http_archive(
    name = "io_bazel_rules_go",
    sha256 = "f87fa87475ea107b3c69196f39c82b7bbf58fe27c62a338684c20ca17d1d8613",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.16.2/rules_go-0.16.2.tar.gz",
)

http_archive(
    name = "com_github_bazelbuild_buildtools",
    strip_prefix = "buildtools-<commit hash>",
    url = "https://github.com/bazelbuild/buildtools/archive/<commit hash>.zip",
)

load("@io_bazel_rules_go//go:def.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@com_github_bazelbuild_buildtools//buildifier:deps.bzl", "buildifier_dependencies")

go_rules_dependencies()

go_register_toolchains()

buildifier_dependencies()
```

`BUILD.bazel` typically in the workspace root:
```bzl
load("@com_github_bazelbuild_buildtools//buildifier:def.bzl", "buildifier")

buildifier(
    name = "buildifier",
)
```
Invoke with
```bash
bazel run //:buildifier
```

## File diagnostics in json

Buildifier supports diagnostics output in machine-readable format (json), triggered by
`--format=json` (only works in combination with `--type=check`). If used in combination with `-v`,
the output json will be indented for better readability.

The output format is the following:

```json
{
    "success": false,  // true if all files are formatted and generate no warnings, false otherwise
    "files": [  // list of all files processed by buildifier
        {
            "filename": "file_1.bzl",
            "formatted": true,  // whether the file is correctly formatted
            "valid": true,  // whether the file is a valid Starlark file. Can only be false if formatted = false
            "warnings": [  // a list of warnings
                {
                    "start": {
                        "line": 1,
                        "column": 5
                    },
                    "end": {
                        "line": 1,
                        "column": 10
                    },
                    "category": "integer-division",
                    "actionable": true,
                    "message": "The \"/\" operator for integer division is deprecated in favor of \"//\".",
                    "url": "https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#integer-division"
                }
            ]
        },
        {
            "filename": "file_2.bzl",
            "formatted": false,
            "valid": true,
            "warnings": [],
            "rewrites": {  // technical information, a list of rewrites buildifier applies during reformatting
                "editoctal": 1
            }
        },
        {
            "filename": "file_3.bzl",
            "formatted": true,
            "valid": true,
            "warnings": []
        },
        {
            "filename": "file_4.not_bzl",
            "formatted": false,
            "valid": false,
            "warnings": []
        }
    ]
}
```

When the `--format` flag is provided, buildifier always returns `0` unless there are internal
failures or wrong input parameters, this means the output can be parsed as JSON, and its `success`
field should be used to determine whether the diagnostics result is positive.
