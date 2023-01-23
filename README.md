gitGPT - Use GPT-3 to Create Git Commit Messages
================================================

gitGPT is a simple Python script that utilizes the [OpenAI GPT-3 API](https://openai.com/blog/openai-api/) to generate git commit messages for you. This script can be especially useful for those who follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for their commit messages.

Installation
------------

To install gitGPT, please follow these steps:

```sh
git clone https://github.com/icereed/gitGPT.git
cd gitGPT
pip install openai
pip install langchain
```

In order to use gitGPT easily, you can create an alias to the `gitGPT.py` script in your `~/.zshrc` file:

```sh
echo "alias gitgpt='python3 $(pwd)/gitGPT.py'" >> ~/.zshrc
source ~/.zshrc
```

Usage
-----

To use gitGPT, you can simply run the following command:

```sh
gitgpt --explain --hint="why did you change this?"
```

The `--explain` flag will generate an explanation for the commit, and the `--hint` flag will provide a hint to GPT-3 on what the commit is about.

Disclaimer
----------

This project is not affiliated with OpenAI or the GPT-3 API. It's just a script that uses the OpenAI GPT-3 API to generate git commit messages.