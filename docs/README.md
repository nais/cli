# Nais CLI manual

This directory contains files related to the [Nais CLI manual](https://cli.nais.io).

The manual is generated from Markdown files in this directory and is available in the Nais CLI documentation. All generated files are ignored by git, and is generated in a GitHub workflow. You can still generate and view the Markdown files locally by running the following command:

```bash
mise run generate:docs
```

You can view the generated HTML documentation locally by using [Jekyll](https://jekyllrb.com/). When installed, run the following commands in this directory:

```bash
mise run generate:docs # Generate the Markdown files
bundle install # Install dependencies from Gemfile
bundle exec jekyll serve # Start the Jekyll server
```

There is also a mise task that does all this in one go:

```bash
mise run docs
```

The command will serve the generated docs on http://localhost:4000.