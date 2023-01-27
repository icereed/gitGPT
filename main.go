package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	gpt3 "github.com/PullRequestInc/go-gpt3"
	"v.io/x/lib/textutil"
)

const (
	API_KEY           = "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"
	REPO_PATH         = "."
	MAX_TOKEN_DAVINCI = 4097
	CACHE_FILE        = "~/.gitgpt/cache"
)

var cache *Cache

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

	if cached, ok := cache.Get(rendered); ok {
		fmt.Println("SummarizeReadme: Using cached result")
		return string(cached), nil
	}

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
	result := cleanString(resp.Choices[0].Text)
	cache.Set(rendered, result)
	return result, nil
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
	if cached, ok := cache.Get(rendered); ok {
		fmt.Println("generateStructureOfRepo: Using cached result")
		return string(cached), nil
	}
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
	result := cleanString(resp.Choices[0].Text)
	cache.Set(rendered, result)
	return result, nil
}

// cleanString removes all non-letter characters except ".", ",", ";" or "!" at the beginning and end of the string
func cleanString(s string) string {
	// Replace all non-letter characters except ".", ",", ";" or "!" at the beginning and end of the string
	re := regexp.MustCompile(`^[^a-zA-Z]*|[^a-zA-Z\.,;!]*$`)
	cleanedString := re.ReplaceAllString(s, "")
	return strings.TrimSpace(cleanedString)
}

func formatGitCommitMessage(commitMessage string) string {
	// io writer into string
	var b bytes.Buffer
	bytesWriter := bufio.NewWriter(&b)

	w := textutil.NewUTF8WrapWriter(bytesWriter, 72)

	r := strings.NewReader(commitMessage)
	if _, err := io.Copy(w, r); err != nil {
		fmt.Println(err)
		return commitMessage
	}

	err := w.Flush()
	if err != nil {
		fmt.Println(err)
		return commitMessage
	}

	bytesWriter.Flush()

	return b.String()
}

func createCommitMessage(llm gpt3.Client, structureOfRepo string, diff string, rawCommitDescription string) (string, error) {
	template := `
I want you to act as an expert software developer.

I will present you a git diff from a commit surrounded by the string "########".
This commit is done in a git repository.

%s

This is the structure of the repository:
%s

Git diff:
########
%s
########

Prompt: "Create a professional commit message describing this change. Keep the description accurate and to the point. Describe also the impact of this change.
The first line must be a summary not longer than 72 characters. Include the detailed description below the title. Use
Conventional Commit messages."
Answer:
`
	// Create commit message here
	if rawCommitDescription != "" {
		rawCommitDescription = fmt.Sprintf("This is the raw commit description: %s", rawCommitDescription)
	}

	tokensWithoutDiff := getNumTokens(fmt.Sprintf(template, rawCommitDescription, structureOfRepo, ""))
	outputTokens := 500

	// Calculate the number of tokens needed for the diff
	tokensLeftForDiff := MAX_TOKEN_DAVINCI - (tokensWithoutDiff + outputTokens)

	// shorten the diff to the number of characters left
	diff = shortenToTokens(diff, tokensLeftForDiff)

	rendered := fmt.Sprintf(template, rawCommitDescription, structureOfRepo, diff)

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
	return formatGitCommitMessage(cleanString(resp.Choices[0].Text)), nil
}

func main() {
	var err error
	cache, err = NewCache(CACHE_FILE)
	defer cache.Write()
	if err != nil {
		log.Fatalln(err)
	}
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
