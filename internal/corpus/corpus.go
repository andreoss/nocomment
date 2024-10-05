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
			Expected: "package main\n\n\nvar s = \"// keep\"\nfunc main() {  }\n",
		},
		{
			Language: "python",
			ID:       "ID-11",
			Input:    "# remove me\ns = \"# keep\"\nx = 1  # trailing\n",
			Expected: "\ns = \"# keep\"\nx = 1  \n",
		},
		{
			Language: "python",
			ID:       "ID-47",
			Input:    "#!/usr/bin/env python3\n# c\nx = 1\n",
			Expected: "#!/usr/bin/env python3\n\nx = 1\n",
		},
		{
			Language: "javascript",
			ID:       "ID-12",
			Input:    "// remove me\nconst s = \"// keep\";\n/* drop */\n",
			Expected: "\nconst s = \"// keep\";\n\n",
		},
		{
			Language: "typescript",
			ID:       "ID-12",
			Input:    "// remove me\nconst s: string = \"// keep\";\n",
			Expected: "\nconst s: string = \"// keep\";\n",
		},
		{
			Language: "java",
			ID:       "ID-13",
			Input:    "// remove me\nString s = \"// keep\";\n/* drop */\n",
			Expected: "\nString s = \"// keep\";\n\n",
		},
		{
			Language: "c",
			ID:       "ID-14",
			Input:    "// remove me\nchar *s = \"// keep\";\n/* drop */\n",
			Expected: "\nchar *s = \"// keep\";\n\n",
		},
		{
			Language: "cpp",
			ID:       "ID-14",
			Input:    "// remove me\nconst char *s = \"// keep\";\n/* drop */\n",
			Expected: "\nconst char *s = \"// keep\";\n\n",
		},
		{
			Language: "rust",
			ID:       "ID-15",
			Input:    "// remove me\nlet s = \"// keep\";\n/* drop */\n",
			Expected: "\nlet s = \"// keep\";\n\n",
		},
		{
			Language: "csharp",
			ID:       "ID-16",
			Input:    "// remove me\nstring s = \"// keep\";\n/* drop */\n",
			Expected: "\nstring s = \"// keep\";\n\n",
		},
		{
			Language: "php",
			ID:       "ID-17",
			Input:    "<?php\n// remove me\n$s = \"// keep\";\n/* drop */\n",
			Expected: "<?php\n\n$s = \"// keep\";\n\n",
		},
		{
			Language: "ruby",
			ID:       "ID-18",
			Input:    "# remove me\ns = \"# keep\"\n",
			Expected: "\ns = \"# keep\"\n",
		},
		{
			Language: "swift",
			ID:       "ID-19",
			Input:    "// remove me\nlet s = \"// keep\"\n/* drop */\n",
			Expected: "\nlet s = \"// keep\"\n\n",
		},
		{
			Language: "kotlin",
			ID:       "ID-20",
			Input:    "// remove me\nval s = \"// keep\"\n/* drop */\n",
			Expected: "\nval s = \"// keep\"\n\n",
		},
		{
			Language: "scala",
			ID:       "ID-21",
			Input:    "// remove me\nval s = \"// keep\"\n/* drop */\n",
			Expected: "\nval s = \"// keep\"\n\n",
		},
		{
			Language: "shell",
			ID:       "ID-22",
			Input:    "# remove me\ns=\"# keep\"\n",
			Expected: "\ns=\"# keep\"\n",
		},
		{
			Language: "sql",
			ID:       "ID-23",
			Input:    "-- remove me\nselect '-- keep';\n/* drop */\n",
			Expected: "\nselect '-- keep';\n\n",
		},
		{
			Language: "html",
			ID:       "ID-24",
			Input:    "<!-- remove me -->\n<p>keep</p>\n",
			Expected: "\n<p>keep</p>\n",
		},
		{
			Language: "jsonc",
			ID:       "ID-25",
			Input:    "// remove me\n{\"k\": \"// keep\"}\n",
			Expected: "\n{\"k\": \"// keep\"}\n",
		},
		{
			Language: "yaml",
			ID:       "ID-26",
			Input:    "# remove me\nkey: \"# keep\"\n",
			Expected: "\nkey: \"# keep\"\n",
		},
		{
			Language: "toml",
			ID:       "ID-27",
			Input:    "# remove me\nkey = \"# keep\"\n",
			Expected: "\nkey = \"# keep\"\n",
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
			Language: "markdown",
			ID:       "ID-28",
			Input:    "<!-- remove me -->\ntext\n",
			Expected: "\ntext\n",
		},
	}
}
