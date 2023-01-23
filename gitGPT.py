import hashlib
import json
import os
import subprocess
import textwrap
from argparse import ArgumentParser, BooleanOptionalAction
from pathlib import Path
from typing import Tuple

from langchain.llms import OpenAI
from langchain.prompts import PromptTemplate

API_KEY = "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"
REPO_PATH = "."


def read_file(file_path: str) -> str:
    with open(file_path, "r") as f:
        return f.read()


def get_directories(path: str) -> str:
    directories = subprocess.check_output(
        f"cd {path}/; find . -type d | grep -v '.git'", shell=True).decode("utf-8")
    return directories


def get_git_diff(path: str) -> str:
    diff = subprocess.check_output(
        f"cd {path}/; git diff --cached", shell=True).decode("utf-8")
    return diff


def summarize_readme(llm, readme: str) -> str:
    template = """
    I want you to act as an expert software developer and product owner.

    I will present you a README.
    The contents of the text are surrounded by the string "######".

    README:
    ######
    {readme}
    ######

    Prompt: "Summarize the contents of the README. Keep the summary short and to the point."
    Answer:
    """

    summarize_readme_prompt = PromptTemplate(
        input_variables=["readme"],
        template=template
    )

    readme_summary = llm(summarize_readme_prompt.format(
        readme=textwrap.shorten(readme, width=6000, placeholder="...")))
    return readme_summary


def describe_structure(llm, readme_summary: str, directories: str) -> str:
    template = """
    I want you to act as an expert software developer and product owner.

    I will present you a summary of the README and a list of directories of a git repository.
    The contents of those texts are surrounded by the string "######".

    Summary of README:
    ######
    {readme_summary}
    ######

    List of directories:
    ######
    {directories}
    ######

    Prompt: "Describe the structure of this repository and what it does. Include a list of frameworks used. Keep the summary short and to the point."
    Answer:
    """

    describe_structure_prompt = PromptTemplate(
        input_variables=["readme_summary", "directories"],
        template=template,
    )

    structure_of_repo = llm(describe_structure_prompt.format(
        readme_summary=readme_summary, directories=directories))
    return structure_of_repo


def format_git_commit_message(commit_message: str) -> str:
    lines = commit_message.splitlines()
    # Keep the first line as is
    commit_message = lines[0] + "\n"

    # Apply textwrap.fill on all other lines
    for line in lines[1:]:
        commit_message += textwrap.fill(line, width=72).strip() + "\n"

    return commit_message.strip()


def create_commit_message(llm, structure_of_repo: str, diff: str, raw_commit_description: str) -> str:
    template = """
    I want you to act as an expert software developer.

    I will present you a git diff from a commit surrounded by the string "########".
    This commit is done in a git repository.

    This is the structure of the repository:
    {structure_of_repo}

    Git diff:
    ########
    {diff}
    ########

    {raw_commit_description}

    Prompt: "Create a professional commit message describing this change. Keep the description accurate and to the point. Describe also the impact of this change.
    The first line must be a summary not longer than 72 characters. Include the detailed description below the title. Use
    Conventional Commit messages."
    Answer:
    """

    # if raw_commit_description is empty, use the default prompt
    if raw_commit_description != "":
        raw_commit_description = template.replace(
            "Raw commit description: \"{raw_commit_description}\"", "")

    commit_message_prompt = PromptTemplate(
        input_variables=["structure_of_repo",
                         "diff", "raw_commit_description"],
        template=template,
    )

    return format_git_commit_message(
        llm(
            commit_message_prompt.format(
                structure_of_repo=structure_of_repo,
                diff=textwrap.shorten(diff, width=5500, placeholder="..."),
                raw_commit_description=raw_commit_description
            )
        )
    )


def read_cache(cache_file: str) -> dict:
    with open(cache_file, "r") as f:
        return json.load(f)


def write_cache(cache_file: str, data: dict):
    os.makedirs(os.path.dirname(cache_file), exist_ok=True)
    with open(cache_file, "w") as f:
        json.dump(data, f)


def get_readme_summary(path: str, cache_file: str, llm) -> str:
    readme = read_file(path)
    readme_hash = hashlib.sha256(readme.encode()).hexdigest()
    if os.path.exists(cache_file):
        try:
            cache = read_cache(cache_file)
            if cache["readme_hash"] == readme_hash:
                readme_summary = cache["readme_summary"]
            else:
                readme_summary = summarize_readme(llm, readme)
                cache["readme_hash"] = readme_hash
                cache["readme_summary"] = readme_summary
                write_cache(cache_file, cache)
        except (FileNotFoundError, json.decoder.JSONDecodeError):
            readme_summary = summarize_readme(llm, readme)
            write_cache(
                cache_file, {"readme_hash": readme_hash, "readme_summary": readme_summary})
    else:
        readme_summary = summarize_readme(llm, readme)
        write_cache(
            cache_file, {"readme_hash": readme_hash, "readme_summary": readme_summary})
    return readme_summary


if __name__ == "__main__":
    os.environ["OPENAI_API_KEY"] = API_KEY
    parser = ArgumentParser()
    parser.add_argument("-m", "--hint", dest="hint", default="")
    parser.add_argument('--explain', action=BooleanOptionalAction)
    args = parser.parse_args()

    # if explain mode is true, print it
    if args.explain:
        print(
            "This is the explain mode.")

    llm = OpenAI(temperature=0.9, model_name="text-davinci-003",
                 max_tokens=500, top_p=1.0, frequency_penalty=0.0,
                 presence_penalty=0.0)

    cache_file = os.path.expanduser("~/.gitgpt/cache.json")
    readme_summary = get_readme_summary(
        Path(REPO_PATH) / "README.md", cache_file, llm)

    directories = get_directories(REPO_PATH)
    structure_of_repo = describe_structure(llm, readme_summary, directories)
    if args.explain:
        print("\n\n--- 🤔 Assuming this context  🤔 ---")
        # print the structure of the repo
        print(structure_of_repo.strip())
    diff = get_git_diff(REPO_PATH)

    # create 3 commit message candidates and choose the best one
    for index in range(3):
        commit_message = create_commit_message(
            llm, structure_of_repo, diff, args.hint)
        print("\n\n--- 📩 Commit message 📩 ---")
        print(commit_message)
        if index != 2:
            print("\n\n--- 🤔 Is this a good commit message? 🤔 ---")
            print("n = no, try again")
            print("y = yes, use this commit message")
            print("y/n?")
            answer = input()
            if answer == "y":
                break

    with open(REPO_PATH + "/.git/gpt_commit", "w") as f:
        f.write(commit_message)
        print("\n\n💾 Commit message written to .git/gpt_commit")
        print("You can now commit with `git commit -F .git/gpt_commit`")
