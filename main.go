package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	gpt3 "github.com/PullRequestInc/go-gpt3"
)

const (
	API_KEY           = "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"
	REPO_PATH         = "."
	MAX_TOKEN_DAVINCI = 4097
)

func readFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	b := make([]byte, 1024)
	var content strings.Builder
	for {
		n, err := file.Read(b)
		if n == 0 || err != nil {
			break
		}
		content.Write(b[:n])
	}
	return content.String(), nil
}

func getDirectories(path string) (string, error) {
	out, err := exec.Command("sh", "-c", fmt.Sprintf("cd %s; git ls-files | xargs dirname | sort | uniq", path)).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func getGitDiff(path string) (string, error) {
	out, err := exec.Command("sh", "-c", fmt.Sprintf("cd %s; git diff --cached", path)).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func summarizeReadme(llm gpt3.Client, readme string) (string, error) {
	template := `
I want you to act as an expert software developer and product owner.

I will present you a README.
The contents of the text are surrounded by the string "######".

README:
######
%s
######

Prompt: "Summarize the contents of the README. Keep the summary short and to the point."
Answer:
`
	// rendered template:
	rendered := fmt.Sprintf(template, readme)
	// Summarize the readme here

	completionRequest := gpt3.CompletionRequest{
		Prompt:      []string{rendered},
		Temperature: gpt3.Float32Ptr(0.9),
		MaxTokens:   gpt3.IntPtr(2000),
		Echo:        false,
	}

	resp, err := llm.Completion(context.Background(), completionRequest)

	if err != nil {
		return "", err
	}

	return cleanString(resp.Choices[0].Text), nil
}

func generateStructureOfRepo(llm gpt3.Client, readmeSummary string, directories string) (string, error) {
	template := `
I want you to act as an expert software developer and product owner.

I will present you a summary of the README and a list of directories of a git repository.
The contents of those texts are surrounded by the string "######".

Summary of README:
######
%s
######

List of directories:
######
%s
######

Prompt: "Describe the structure of this repository and what it does. Include a list of frameworks used. Keep the summary short and to the point."
Answer:
`
	// Generate structure of repo here
	rendered := fmt.Sprintf(template, readmeSummary, directories)
	completionRequest := gpt3.CompletionRequest{
		Prompt:      []string{rendered},
		Temperature: gpt3.Float32Ptr(0.9),
		MaxTokens:   gpt3.IntPtr(2000),
		Echo:        false,
	}
	resp, err := llm.Completion(context.Background(), completionRequest)
	if err != nil {
		return "", err
	}
	return cleanString(resp.Choices[0].Text), nil
}

// cleanString removes all non-letter characters except ".", ",", ";" or "!" at the beginning and end of the string
func cleanString(s string) string {
	// Replace all non-letter characters except ".", ",", ";" or "!" at the beginning and end of the string
	re := regexp.MustCompile(`^[^a-zA-Z]*|[^a-zA-Z.,;!]*$`)
	cleanedString := re.ReplaceAllString(s, "")
	return strings.TrimSpace(cleanedString)
}

func formatGitCommitMessage(commitMessage string) string {
	// Format the commit message here
	// Make sure the first line is not longer than 72 characters
	// Make sure the description is not longer than 80 characters
	// Make sure the commit message is separated into a title and a description
	// Make sure the title is written in imperative mood
	// Make sure the title is written in the present tense
	// Make sure the title is written in the active voice
	// Make sure the title is written in the Conventional Commit format
	// Make sure the description is written in the Conventional Commit format
	return commitMessage
}

func createCommitMessage(llm gpt3.Client, structureOfRepo string, diff string, rawCommitDescription string) (string, error) {
	template := `
I want you to act as an expert software developer.

I will present you a git diff from a commit surrounded by the string "########" and you will create a git commit based on a prompt.
Create a professional commit message describing this change.
Keep the description accurate and to the point.
Describe also the impact of this change.
Make sure the first line is not longer than 72 characters
Make sure the description is not longer than 80 characters
Make sure the commit message is separated into a title and a description
Make sure the title is written in imperative mood
Make sure the title is written in the present tense
Make sure the title is written in the active voice
Make sure the title is written in the Conventional Commit format
Make sure the description is written in the Conventional Commit format

This commit is done in a git repository.

%s

This is the structure of the repository:
%s

Git diff:
########
%s
########

Answer:
`
	// Create commit message here
	if rawCommitDescription != "" {
		rawCommitDescription = fmt.Sprintf("This is the raw commit description: %s", rawCommitDescription)
	}

	tokensWithoutDiff := getNumTokens(fmt.Sprintf(template, rawCommitDescription, structureOfRepo, ""))
	fmt.Printf("Tokens without diff: %d\n", tokensWithoutDiff)

	outputTokens := 500

	// Calculate the number of tokens needed for the diff

	tokensLeftForDiff := MAX_TOKEN_DAVINCI - (tokensWithoutDiff + outputTokens)

	charactersLeftForDiff := getNumChars(tokensLeftForDiff)

	fmt.Printf("Tokens left for diff: %d\n", tokensLeftForDiff)
	fmt.Printf("Len of diff: %d\n", len(diff))

	// shorten the diff to the number of characters left
	if len(diff) > charactersLeftForDiff {
		diff = diff[:charactersLeftForDiff]
	}

	fmt.Printf("Len of shortened diff: %d\n", len(diff))

	rendered := fmt.Sprintf(template, rawCommitDescription, structureOfRepo, diff)
	tokensWithDiff := getNumTokens(rendered)
	fmt.Printf("Tokens with diff: %d\n", tokensWithDiff)

	completionRequest := gpt3.CompletionRequest{
		Prompt:      []string{rendered},
		Temperature: gpt3.Float32Ptr(0.9),
		MaxTokens:   gpt3.IntPtr(outputTokens),
		Echo:        false,
	}
	resp, err := llm.Completion(context.Background(), completionRequest)
	if err != nil {
		return "", err
	}
	return cleanString(resp.Choices[0].Text), nil
}

func main() {
	// ctx := context.Background()
	llm := gpt3.NewClient(API_KEY, gpt3.WithDefaultEngine(gpt3.TextDavinci003Engine))

	readme, err := readFile(filepath.Join(REPO_PATH, "README.md"))
	if err != nil {
		fmt.Println(err)
		return
	}

	directories, err := getDirectories(REPO_PATH)
	if err != nil {
		fmt.Println(err)
		return
	}

	diff, err := getGitDiff(REPO_PATH)
	if err != nil {
		fmt.Println(err)
		return
	}

	readmeSummary, err := summarizeReadme(llm, readme)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("-------------\nSummary of README:\n%s\n", readmeSummary)

	structureOfRepo, err := generateStructureOfRepo(llm, readmeSummary, directories)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("-------------\nStructure of repo:\n%s\n", structureOfRepo)

	commitMessage, err := createCommitMessage(llm, structureOfRepo, diff, "")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("-------------\nCommit message:\n%s\n", commitMessage)
}
