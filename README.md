gitGPT - Use GPT-3 to Create Git Commit Messages
================================================

gitGPT is a go CLI app that utilizes the [OpenAI GPT-3 API](https://openai.com/blog/openai-api/) to generate git commit messages for you. This app follows the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for the commit messages.

Installation
------------

```sh
go get -u github.com/icereed/gitGPT
```

Usage
-----

To use gitGPT, you can simply run the following command:

```sh
export OPENAI_API_KEY=<your api key>

gitgpt --help
A tool for summarizing a Git repository using GPT-3, including the README, directory structure, and commit message

Usage:
  gitgpt [flags]

Flags:
  -e, --explain       Turn on console output for intermediate results
  -h, --help          help for gitgpt
      --hint string   Provide a hint for the commit message


gitgpt --explain --hint="why did you change this?"
```

The `--explain` flag will generate an explanation for the commit, and the `--hint` flag will provide a hint to GPT-3 on what the commit is about.

Disclaimer
----------

This project is not affiliated with OpenAI or the GPT-3 API. It's just a script that uses the OpenAI GPT-3 API to generate git commit messages.