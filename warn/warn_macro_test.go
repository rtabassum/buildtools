package warn

import "testing"

func TestUnnamedMacroNoReaderSameFile(t *testing.T) {
	checkFindings(t, "unnamed-macro", `
load(":foo.bzl", "foo")

my_rule = rule()

def macro(x):
  foo()
  my_rule(name = x)

def not_macro(x):
  foo()
  native.glob()
  native.existing_rule()
  native.existing_rules()
  return my_rule

def another_macro(x):
  foo()
  [native.cc_library() for i in x]
`,
		[]string{
			`5: Macro function "macro" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "my_rule" on line 7.`,
			`16: Macro function "another_macro" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "native.cc_library" on line 18.`,
		},
		scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro1(foo, name, bar):
  my_rule()

def macro2(foo, *, name):
  my_rule()

def macro3(foo, *args, **kwargs):
  my_rule()
`,
		[]string{},
		scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro(name):
  my_rule(name = name)

alias = macro

def bad_macro():
  for x in y:
    alias(x)
`,
		[]string{
			`8: Macro function "bad_macro" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "alias" on line 10.`,
		},
		scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro1():
  my_rule(name = x)

def macro2():
  macro1()

def macro3():
  macro2()

def macro4():
  my_rule()
`,
		[]string{
			`3: Macro function "macro1" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "my_rule" on line 4.`,
			`6: Macro function "macro2" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "macro1" on line 7`,
			`9: Macro function "macro3" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "macro2" on line 10.`,
			`12: Macro function "macro4" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "my_rule" on line 13.`,
		},
		scopeBzl)
}

func TestUnnamedMacroRecursion(t *testing.T) {
	// Recursion is not allowed in Bazel, but shouldn't cause Buildifier to crash

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro():
  macro()
`, []string{}, scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro():
  macro()
`, []string{}, scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro1():
  macro2()

def macro2():
  macro3()

def macro3():
  macro1()
`, []string{}, scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def foo():
  bar()

def bar():
  foo()
  my_rule()
`,
		[]string{
			`3: Macro function "foo" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "bar" on line 4.`,
			`6: Macro function "bar" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "my_rule" on line 8.`,
		},
		scopeBzl)
}

func TestUnnamedMacroWithReader(t *testing.T) {
	defer setUpFileReader(map[string]string{
		"test/package/foo.bzl": `
def foo():
  native.foo_binary()

def bar():
  foo()

my_rule = rule()
`,
		"test/package/baz.bzl": `
load(":foo.bzl", "bar", your_rule = "my_rule")
load("//does/not:exist.bzl", "something")

def baz():
  if False:
    bar()

def qux():
  your_rule()

def f():
  something()
`,
	})()

	checkFindings(t, "unnamed-macro", `
load("//test/package/foo.bzl", abc = "bar")
load(":baz.bzl", "baz", "qux", "f")

def macro1(surname):
  abc()

def macro2(surname):
  baz()

def macro3(surname):
  qux()

def not_macro(x):
  f()
`,
		[]string{
			`4: Macro function "macro1" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "abc" on line 5.`,
			`7: Macro function "macro2" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "baz" on line 8.`,
			`10: Macro function "macro3" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "qux" on line 11.`,
		},
		scopeBzl)
}

func TestUnnamedMacroRecursionWithReader(t *testing.T) {
	// Recursion is not allowed in Bazel, but shouldn't cause Buildifier to crash

	defer setUpFileReader(map[string]string{
		"test/package/foo.bzl": `
load(":bar.bzl", "bar")

def foo():
  foo()

def baz():
  bar()

def qux():
  native.cc_library()

`,
		"test/package/bar.bzl": `
load(":foo.bzl", "foo", "baz", quuux = "qux")

def bar():
  baz()

def qux():
  foo()
  baz()
  quuux()
`,
	})()

	checkFindings(t, "unnamed-macro", `
load(":foo.bzl", "foo", "baz")
load(":bar.bzl", quux = "qux")

def macro():
  foo()
  baz()
  quux()
`, []string{
		`4: Macro function "macro" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "quux" on line 7.`,
	}, scopeBzl)
}

func TestUnnamedMacroLoadCycle(t *testing.T) {
	defer setUpFileReader(map[string]string{
		"test/package/foo.bzl": `
load(":test_file.bzl", some_rule = "my_rule")

def foo():
  some_rule()
`,
	})()

	checkFindings(t, "unnamed-macro", `
load(":foo.bzl", bar = "foo")

my_rule = rule()

def macro():
  bar()
`, []string{
		`5: Macro function "macro" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "bar" on line 6.`,
	}, scopeBzl)
}

func TestUnnamedMacroLoadedFiles(t *testing.T) {
	// Test that not necessary files are not loaded

	defer setUpFileReader(map[string]string{
		"a.bzl": "a = rule()",
		"b.bzl": "b = rule()",
		"c.bzl": "c = rule()",
		"d.bzl": "d = rule()",
	})()

	checkFindings(t, "unnamed-macro", `
load("//a.bzl", "a")
load("//b.bzl", "b")
load("//c.bzl", "c")
load("//d.bzl", "d")

def macro1():
  a()  # has to load a.bzl to analyze

def macro2():
  b()  # can skip b.bzl because there's a native rule
  native.cc_library()

def macro3():
  c()  # can skip c.bzl because a.bzl has already been loaded
  a()

def macro4():
  d()  # can skip d.bzl because there's a rule or another macro defined in the same file
  r()

r = rule()
`, []string{
		`6: Macro function "macro1" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "a" on line 7.`,
		`9: Macro function "macro2" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "native.cc_library" on line 11.`,
		`13: Macro function "macro3" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "a" on line 15.`,
		`17: Macro function "macro4" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "r" on line 19.`,
	}, scopeBzl)

	if len(fileReaderRequests) == 1 && fileReaderRequests[0] == "a.bzl" {
		return
	}
	t.Errorf("expected to load only a.bzl, instead loaded %d files: %v", len(fileReaderRequests), fileReaderRequests)
}

func TestUnnamedMacroAliases(t *testing.T) {
	// Test that not necessary files are not loaded

	defer setUpFileReader(map[string]string{
		"test/package/foo.bzl": `
load(":bar.bzl", _bar = "bar")

bar = _bar`,
		"test/package/bar.bzl": `
my_rule = rule()

bar = my_rule`,
	})()

	checkFindings(t, "unnamed-macro", `
load(":bar.bzl", "bar")

baz = bar

def macro1():
  baz()

def macro2(name):
  baz()
`, []string{
		`5: Macro function "macro1" doesn't accept a keyword argument "name".

It is considered a macro because it calls a rule or another macro "baz" on line 6.`,
	}, scopeBzl)
}
