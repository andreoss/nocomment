package corpus

type Case struct {
	Language string
	ID       string
	Input    string
	Expected string
}

func Cases() []Case {
	return []Case{
		{
			Language: "go",
			ID:       "ID-10",
			Input:    "package main\n\n// remove me\nvar s = \"// keep\"\nfunc main() { /* drop */ }\n",
			Expected: "package main\n\nvar s = \"// keep\"\nfunc main() {  }\n",
		},
		{
			Language: "python",
			ID:       "ID-11",
			Input:    "# remove me\ns = \"# keep\"\nx = 1  # trailing\n",
			Expected: "s = \"# keep\"\nx = 1  \n",
		},
		{
			Language: "python",
			ID:       "ID-47",
			Input:    "#!/usr/bin/env python3\n# c\nx = 1\n",
			Expected: "#!/usr/bin/env python3\nx = 1\n",
		},
		{
			Language: "javascript",
			ID:       "ID-12",
			Input:    "// remove me\nconst s = \"// keep\";\n/* drop */\n",
			Expected: "const s = \"// keep\";\n",
		},
		{
			Language: "typescript",
			ID:       "ID-12",
			Input:    "// remove me\nconst s: string = \"// keep\";\n",
			Expected: "const s: string = \"// keep\";\n",
		},
		{
			Language: "java",
			ID:       "ID-13",
			Input:    "// remove me\nString s = \"// keep\";\n/* drop */\n",
			Expected: "String s = \"// keep\";\n",
		},
		{
			Language: "c",
			ID:       "ID-14",
			Input:    "// remove me\nchar *s = \"// keep\";\n/* drop */\n",
			Expected: "char *s = \"// keep\";\n",
		},
		{
			Language: "cpp",
			ID:       "ID-14",
			Input:    "// remove me\nconst char *s = \"// keep\";\n/* drop */\n",
			Expected: "const char *s = \"// keep\";\n",
		},
		{
			Language: "rust",
			ID:       "ID-15",
			Input:    "// remove me\nlet s = \"// keep\";\n/* drop */\n",
			Expected: "let s = \"// keep\";\n",
		},
		{
			Language: "csharp",
			ID:       "ID-16",
			Input:    "// remove me\nstring s = \"// keep\";\n/* drop */\n",
			Expected: "string s = \"// keep\";\n",
		},
		{
			Language: "php",
			ID:       "ID-17",
			Input:    "<?php\n// remove me\n$s = \"// keep\";\n/* drop */\n",
			Expected: "<?php\n$s = \"// keep\";\n",
		},
		{
			Language: "ruby",
			ID:       "ID-18",
			Input:    "# remove me\ns = \"# keep\"\n",
			Expected: "s = \"# keep\"\n",
		},
		{
			Language: "swift",
			ID:       "ID-19",
			Input:    "// remove me\nlet s = \"// keep\"\n/* drop */\n",
			Expected: "let s = \"// keep\"\n",
		},
		{
			Language: "kotlin",
			ID:       "ID-20",
			Input:    "// remove me\nval s = \"// keep\"\n/* drop */\n",
			Expected: "val s = \"// keep\"\n",
		},
		{
			Language: "scala",
			ID:       "ID-21",
			Input:    "// remove me\nval s = \"// keep\"\n/* drop */\n",
			Expected: "val s = \"// keep\"\n",
		},
		{
			Language: "shell",
			ID:       "ID-22",
			Input:    "# remove me\ns=\"# keep\"\n",
			Expected: "s=\"# keep\"\n",
		},
		{
			Language: "sql",
			ID:       "ID-23",
			Input:    "-- remove me\nselect '-- keep';\n/* drop */\n",
			Expected: "select '-- keep';\n",
		},
		{
			Language: "postgresql",
			ID:       "ID-55",
			Input:    "-- remove me\nselect '-- keep';\n/* drop */\n",
			Expected: "select '-- keep';\n",
		},
		{
			Language: "html",
			ID:       "ID-24",
			Input:    "<!-- remove me -->\n<p>keep</p>\n",
			Expected: "<p>keep</p>\n",
		},
		{
			Language: "xml",
			ID:       "ID-48",
			Input:    "<?xml version=\"1.0\"?>\n<!-- remove me -->\n<root>keep</root>\n",
			Expected: "<?xml version=\"1.0\"?>\n<root>keep</root>\n",
		},
		{
			Language: "jsonc",
			ID:       "ID-25",
			Input:    "// remove me\n{\"k\": \"// keep\"}\n",
			Expected: "{\"k\": \"// keep\"}\n",
		},
		{
			Language: "yaml",
			ID:       "ID-26",
			Input:    "# remove me\nkey: \"# keep\"\n",
			Expected: "key: \"# keep\"\n",
		},
		{
			Language: "toml",
			ID:       "ID-27",
			Input:    "# remove me\nkey = \"# keep\"\n",
			Expected: "key = \"# keep\"\n",
		},
		{
			Language: "javascript",
			ID:       "ID-12",
			Input:    "const a = `// not`; // drop\n",
			Expected: "const a = `// not`; \n",
		},
		{
			Language: "rust",
			ID:       "ID-15",
			Input:    "let s = 1; /* a /* b */ c */\nlet t = 2;\n",
			Expected: "let s = 1; \nlet t = 2;\n",
		},
		{
			Language: "scala",
			ID:       "ID-21",
			Input:    "val s = 1 /* a /* b */ c */\nval t = 2\n",
			Expected: "val s = 1 \nval t = 2\n",
		},
		{
			Language: "javascript",
			ID:       "ID-50",
			Input:    "const re = /a\\/\\//; // drop\n",
			Expected: "const re = /a\\/\\//; \n",
		},
		{
			Language: "ruby",
			ID:       "ID-51",
			Input:    "s = %q{# not a comment}\n# drop\n",
			Expected: "s = %q{# not a comment}\n",
		},
		{
			Language: "shell",
			ID:       "ID-52",
			Input:    "cat <<EOF\n# not a comment\nEOF\n# drop\n",
			Expected: "cat <<EOF\n# not a comment\nEOF\n",
		},
		{
			Language: "php",
			ID:       "ID-53",
			Input:    "<?php\n$s = <<<EOT\n# not a comment\nEOT;\n# drop\n",
			Expected: "<?php\n$s = <<<EOT\n# not a comment\nEOT;\n",
		},
		{
			Language: "yaml",
			ID:       "ID-54",
			Input:    "key: |\n  # not a comment\n# drop\n",
			Expected: "key: |\n  # not a comment\n",
		},
		{
			Language: "markdown",
			ID:       "ID-28",
			Input:    "<!-- remove me -->\ntext\n",
			Expected: "text\n",
		},
	}
}
