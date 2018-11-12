package warn

import (
	"testing"
)

func TestAttrConfigurationWarning(t *testing.T) {
	checkFindingsAndFix(t, "attr-cfg", `
rule(
  attrs = {
      "foo": attr.label_list(mandatory = True, cfg = "data"),
  }
)

attr.label_list(mandatory = True, cfg = "host")`, `
rule(
  attrs = {
      "foo": attr.label_list(mandatory = True),
  }
)

attr.label_list(mandatory = True, cfg = "host")`,
		[]string{":3: cfg = \"data\" for attr definitions has no effect and should be removed."},
		false)
}

func TestAttrNonEmptyWarning(t *testing.T) {
	checkFindingsAndFix(t, "attr-non-empty", `
rule(
  attrs = {
      "foo": attr.label_list(mandatory = True, non_empty = True),
      "bar": attr.label_list(mandatory = True, non_empty = False),
      "baz": attr.label_list(mandatory = True, non_empty = foo.bar()),
      "qux": attr.label_list(mandatory = True, non_empty = not foo.bar()),
      "aaa": attr.label_list(mandatory = True, non_empty = (foo.bar())),
      "bbb": attr.label_list(mandatory = True, non_empty = (
					not foo.bar()
			)),
  }
)`, `
rule(
  attrs = {
      "foo": attr.label_list(mandatory = True, allow_empty = False),
      "bar": attr.label_list(mandatory = True, allow_empty = True),
      "baz": attr.label_list(mandatory = True, allow_empty = not foo.bar()),
      "qux": attr.label_list(mandatory = True, allow_empty = foo.bar()),
      "aaa": attr.label_list(mandatory = True, allow_empty = (not foo.bar())),
      "bbb": attr.label_list(mandatory = True, allow_empty = (
					foo.bar()
			)),
  }
)`,
		[]string{
			":3: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
			":4: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
			":5: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
			":6: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
			":7: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
			":8: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
		},
		false)
}

func TestAttrSingleFileWarning(t *testing.T) {
	checkFindingsAndFix(t, "attr-single-file", `
rule(
  attrs = {
      "foo": attr.label_list(single_file = True, allow_files = [".cc"], mandatory = True),
      "bar": attr.label_list(single_file = True, mandatory = True),
      "baz": attr.label_list(single_file = False, mandatory = True),
  }
)`, `
rule(
  attrs = {
      "foo": attr.label_list(allow_single_file = [".cc"], mandatory = True),
      "bar": attr.label_list(allow_single_file = True, mandatory = True),
      "baz": attr.label_list(mandatory = True),
	}
)`,
		[]string{
			":3: single_file is deprecated in favor of allow_single_file.",
			":4: single_file is deprecated in favor of allow_single_file.",
			":5: single_file is deprecated in favor of allow_single_file.",
		},
		false)
}

func TestCtxActionsWarning(t *testing.T) {
	checkFindingsAndFix(t, "ctx-actions", `
def impl(ctx):
  ctx.new_file(foo)
  ctx.new_file(foo, bar)
  ctx.new_file(foo, bar, baz)
  ctx.experimental_new_directory(foo, bar)
  ctx.file_action(foo, bar)
  ctx.file_action(foo, bar, executable = True)
  ctx.action(foo, bar, command = "foo")
  ctx.action(foo, bar, executable = "bar")
  ctx.empty_action(foo, bar)
  ctx.template_action(foo, bar)
  ctx.template_action(foo, bar, executable = True)
	ctx.foobar(foo, bar)
`, `
def impl(ctx):
  ctx.actions.declare_file(foo)
  ctx.actions.declare_file(bar, sibling = foo)
  ctx.new_file(foo, bar, baz)
  ctx.actions.declare_directory(foo, bar)
  ctx.actions.write(foo, bar)
  ctx.actions.write(foo, bar, is_executable = True)
  ctx.actions.run_shell(foo, bar, command = "foo")
  ctx.actions.run(foo, bar, executable = "bar")
  ctx.actions.do_nothing(foo, bar)
  ctx.actions.expand_template(foo, bar)
  ctx.actions.expand_template(foo, bar, is_executable = True)
	ctx.foobar(foo, bar)
`,
		[]string{
			":2: \"ctx.new_file\" is deprecated.",
			":3: \"ctx.new_file\" is deprecated.",
			":4: \"ctx.new_file\" is deprecated.",
			":5: \"ctx.experimental_new_directory\" is deprecated.",
			":6: \"ctx.file_action\" is deprecated.",
			":7: \"ctx.file_action\" is deprecated.",
			":8: \"ctx.action\" is deprecated.",
			":9: \"ctx.action\" is deprecated.",
			":10: \"ctx.empty_action\" is deprecated.",
			":11: \"ctx.template_action\" is deprecated.",
			":12: \"ctx.template_action\" is deprecated.",
		},
		false)

	checkFindings(t, "ctx-actions", `
def impl(ctx):
  ctx.new_file(foo, bar, baz)
`, []string{
		":2: \"ctx.new_file\" is deprecated.",
	}, false)
}

func TestPackageNameWarning(t *testing.T) {
	checkFindingsAndFix(t, "package-name", `
foo(a = PACKAGE_NAME)

def f(PACKAGE_NAME):
    foo(a = PACKAGE_NAME)

def g():
    foo(a = PACKAGE_NAME)
`, `
foo(a = native.package_name())

def f(PACKAGE_NAME):
    foo(a = PACKAGE_NAME)

def g():
    foo(a = native.package_name())
`,
		[]string{
			":1: Global variable \"PACKAGE_NAME\" is deprecated in favor of \"native.package_name()\". Please rename it.",
			":7: Global variable \"PACKAGE_NAME\" is deprecated in favor of \"native.package_name()\". Please rename it.",
		},
		false)

	checkFindings(t, "package-name", `
PACKAGE_NAME = "foo"
foo(a = PACKAGE_NAME)
`, []string{}, false)

	checkFindings(t, "package-name", `
load(":foo.bzl", "PACKAGE_NAME")
foo(a = PACKAGE_NAME)
`, []string{}, false)
}

func TestRepositoryNameWarning(t *testing.T) {
	checkFindingsAndFix(t, "repository-name", `
foo(a = REPOSITORY_NAME)

def f(REPOSITORY_NAME):
    foo(a = REPOSITORY_NAME)

def g():
    foo(a = REPOSITORY_NAME)
`, `
foo(a = native.repository_name())

def f(REPOSITORY_NAME):
    foo(a = REPOSITORY_NAME)

def g():
    foo(a = native.repository_name())
`,
		[]string{
			":1: Global variable \"REPOSITORY_NAME\" is deprecated in favor of \"native.repository_name()\". Please rename it.",
			":7: Global variable \"REPOSITORY_NAME\" is deprecated in favor of \"native.repository_name()\". Please rename it.",
		}, false)

	checkFindings(t, "repository-name", `
REPOSITORY_NAME = "foo"
foo(a = REPOSITORY_NAME)
`, []string{}, false)

	checkFindings(t, "repository-name", `
load(":foo.bzl", "REPOSITORY_NAME")
foo(a = REPOSITORY_NAME)
`, []string{}, false)
}

func TestFileTypeNameWarning(t *testing.T) {
	checkFindings(t, "filetype", `
rule1(types=FileType([".cc", ".h"]))
rule2(types=FileType(types=[".cc", ".h"]))

FileType(foobar)

def macro1():
    a = FileType([".py"])

def macro2():
    FileType = foo
    b = FileType([".java"])
`, []string{
		":1: The FileType function is deprecated.",
		":2: The FileType function is deprecated.",
		":4: The FileType function is deprecated.",
		":7: The FileType function is deprecated.",
	}, false)

	checkFindings(t, "filetype", `
FileType = foo

rule1(types=FileType([".cc", ".h"]))
rule2(types=FileType(types=[".cc", ".h"]))
`, []string{}, false)
}

func TestOutputGroupWarning(t *testing.T) {
	checkFindingsAndFix(t, "output-group", `
def _impl(ctx):
    bin = ctx.attr.my_dep.output_group.bin
`, `
def _impl(ctx):
    bin = ctx.attr.my_dep[OutputGroupInfo].bin
`,
		[]string{
			":2: \"ctx.attr.dep.output_group\" is deprecated in favor of \"ctx.attr.dep[OutputGroupInfo]\".",
		},
		false)
}

func TestNativeGitRepositoryWarning(t *testing.T) {
	checkFindingsAndFix(t, "git-repository", `
"""My file"""

def macro():
    git_repository(foo, bar)
`, `
"""My file"""

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

def macro():
    git_repository(foo, bar)
`,
		[]string{
			":4: Function \"git_repository\" is not global anymore and needs to be loaded from \"@bazel_tools//tools/build_defs/repo:git.bzl\".",
		},
		false)

	checkFindingsAndFix(t, "git-repository", `
"""My file"""

def macro():
    git_repository(foo, bar)
    new_git_repository(foo, bar)
`, `
"""My file"""

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository", "new_git_repository")

def macro():
    git_repository(foo, bar)
    new_git_repository(foo, bar)
`,
		[]string{
			":4: Function \"git_repository\" is not global anymore and needs to be loaded from \"@bazel_tools//tools/build_defs/repo:git.bzl\".",
			":5: Function \"new_git_repository\" is not global anymore and needs to be loaded from \"@bazel_tools//tools/build_defs/repo:git.bzl\".",
		},
		false)

	checkFindingsAndFix(t, "git-repository", `
"""My file"""

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

def macro():
    git_repository(foo, bar)
    new_git_repository(foo, bar)
`, `
"""My file"""

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository", "new_git_repository")

def macro():
    git_repository(foo, bar)
    new_git_repository(foo, bar)
`,
		[]string{
			":7: Function \"new_git_repository\" is not global anymore and needs to be loaded from \"@bazel_tools//tools/build_defs/repo:git.bzl\".",
		},
		false)
}

func TestNativeHttpArchiveWarning(t *testing.T) {
	checkFindingsAndFix(t, "http-archive", `
"""My file"""

def macro():
    http_archive(foo, bar)
`, `
"""My file"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def macro():
    http_archive(foo, bar)
`,
		[]string{
			":4: Function \"http_archive\" is not global anymore and needs to be loaded from \"@bazel_tools//tools/build_defs/repo:http.bzl\".",
		},
		false)
}

func TestArgsApiWarning(t *testing.T) {
	checkFindingsAndFix(t, "args-api", `
def impl(ctx):
    args = ctx.actions.args()
    args.add(foo, bar)
    args.add(foo, bar, before_each = aaa)
    args.add(foo, bar, join_with = bbb)
    args.add(foo, bar, before_each = ccc, join_with = ddd)
    args.add(foo, bar, map_fn = eee)
    args.add(foo, bar, map_fn = fff, before_each = ggg)
    args.add(foo, bar, map_fn = hhh, join_with = iii)
    args.add(foo, bar, map_fn = jjj, before_each = kkk, join_with = lll)
`, `
def impl(ctx):
    args = ctx.actions.args()
    args.add(foo, bar)
    args.add_all(foo, bar, before_each = aaa)
    args.add_joined(foo, bar, join_with = bbb)
    args.add_joined(foo, bar, format_each = ccc + "%s", join_with = ddd)
    args.add_all(foo, bar, map_each = eee)
    args.add_all(foo, bar, map_each = fff, before_each = ggg)
    args.add_joined(foo, bar, map_each = hhh, join_with = iii)
    args.add_joined(foo, bar, map_each = jjj, format_each = kkk + "%s", join_with = lll)
`,
		[]string{
			":4: \"ctx.actions.args().add()\" for multiple arguments is deprecated in favor of \"add_all()\" or \"add_joined()\".",
			":5: \"ctx.actions.args().add()\" for multiple arguments is deprecated in favor of \"add_all()\" or \"add_joined()\".",
			":6: \"ctx.actions.args().add()\" for multiple arguments is deprecated in favor of \"add_all()\" or \"add_joined()\".",
			":7: \"ctx.actions.args().add()\" for multiple arguments is deprecated in favor of \"add_all()\" or \"add_joined()\".",
			":8: \"ctx.actions.args().add()\" for multiple arguments is deprecated in favor of \"add_all()\" or \"add_joined()\".",
			":9: \"ctx.actions.args().add()\" for multiple arguments is deprecated in favor of \"add_all()\" or \"add_joined()\".",
			":10: \"ctx.actions.args().add()\" for multiple arguments is deprecated in favor of \"add_all()\" or \"add_joined()\".",
		},
		false)
}