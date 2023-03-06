package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	gpt3 "github.com/PullRequestInc/go-gpt3"
	"github.com/spf13/cobra"
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
	completionRequest := gpt3.ChatCompletionRequest{
		Messages: []gpt3.ChatCompletionRequestMessage{
			{
				Role:    "user",
				Content: rendered,
			},
		},
		Temperature: 0.9,
		MaxTokens:   2000,
	}
	resp, err := llm.ChatCompletion(context.Background(), completionRequest)
	if err != nil {
		return "", err
	}
	result := cleanString(resp.Choices[0].Message.Content)
	cache.Set(rendered, result)
	return result, nil
}

func summarizeDiff(llm gpt3.Client, diff string) (string, error) {
	template := `
I want you to act as an expert software developer. I will present you a git diff.
Your job is to explain the change file by file. Keep the explanation short and to the point.

git diff:
########
%s
########
`
	template = strings.TrimSpace(template)
	tokensWithoutDiff := getNumTokens(fmt.Sprintf(template, ""))
	outputTokens := 500

	// Calculate the number of tokens needed for the diff
	tokensLeftForDiff := MAX_TOKEN_DAVINCI - (tokensWithoutDiff + outputTokens)

	// shorten the diff to the number of characters left
	diff = shortenToTokens(diff, tokensLeftForDiff)

	rendered := fmt.Sprintf(template, diff)
	if cached, ok := cache.Get(rendered); ok {
		fmt.Println("summarizeDiff: Using cached result")
		return string(cached), nil
	}
	completionRequest := gpt3.ChatCompletionRequest{
		Messages: []gpt3.ChatCompletionRequestMessage{
			{
				Role:    "user",
				Content: rendered,
			},
		},
		Temperature: 0.9,
		MaxTokens:   outputTokens,
	}
	resp, err := llm.ChatCompletion(context.Background(), completionRequest)
	if err != nil {
		return "", err
	}
	result := cleanString(resp.Choices[0].Message.Content)
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
I want you to act as an expert software developer. Your job is to create a git commit message based on the input below.
%s

This is the structure of the git repository:
START STRUCTURE
%s
END STRUCTURE

Changes:
START CHANGES
%s
END CHANGES

Create a professional commit message describing this change. Keep the description accurate and to the point. Describe also the impact of this change. The first line must be a summary not longer than 72 characters. Include the detailed description below the title. Don't include PR links. Use Conventional Commit messages.
`
	template = strings.TrimSpace(template)
	// Create commit message here
	if rawCommitDescription != "" {
		rawCommitDescription = fmt.Sprintf("This is the raw commit description: \"%s\"", rawCommitDescription)
	}

	tokensWithoutDiff := getNumTokens(fmt.Sprintf(template, rawCommitDescription, structureOfRepo, ""))
	outputTokens := 500

	// Calculate the number of tokens needed for the diff
	tokensLeftForDiff := MAX_TOKEN_DAVINCI - (tokensWithoutDiff + outputTokens)

	// shorten the diff to the number of characters left
	diff = shortenToTokens(diff, tokensLeftForDiff)

	rendered := fmt.Sprintf(template, rawCommitDescription, structureOfRepo, diff)

	data := []byte(rendered)
	err := ioutil.WriteFile("prompt.txt", data, 0666) // For debugging

	if err != nil {
		log.Fatal(err)
	}
	completionRequest := gpt3.ChatCompletionRequest{
		Messages: []gpt3.ChatCompletionRequestMessage{
			{
				Role:    "user",
				Content: rendered,
			},
		},
		Temperature: 0.9,
		MaxTokens:   outputTokens,
	}
	resp, err := llm.ChatCompletion(context.Background(), completionRequest)
	if err != nil {
		return "", err
	}
	return formatGitCommitMessage(cleanString(resp.Choices[0].Message.Content)), nil
}

func main() {
	var err error
	explain := false
	hint := ""
	rootCmd := &cobra.Command{
		Use:   "gitgpt",
		Short: "A tool for summarizing a Git repository using GPT-3",
		Long:  "A tool for summarizing a Git repository using GPT-3, including the README, directory structure, and commit message",
		Run: func(cmd *cobra.Command, args []string) {
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

			if explain {
				fmt.Printf("\nSummary of README:\n%s\n-------------\n", readmeSummary)
			}

			structureOfRepo, err := generateStructureOfRepo(llm, readmeSummary, directories)
			if err != nil {
				fmt.Println(err)
				return
			}
			if explain {
				fmt.Printf("\nStructure of repo:\n%s\n-------------\n", structureOfRepo)
			}

			// diff summary
			diffSummary, err := summarizeDiff(llm, diff)
			if err != nil {
				fmt.Println(err)
				return
			}

			if explain {
				fmt.Printf("\nSummary of diff:\n%s\n-------------\n", diffSummary)
			}

			commitMessage, err := createCommitMessage(llm, structureOfRepo, diffSummary, hint)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("\nCommit message:\n%s\n-------------\n", commitMessage)

			// Write the commit message to a file: .git/gpt_commit

			commitMessageFile := filepath.Join(REPO_PATH, ".git", "gpt_commit")
			err = ioutil.WriteFile(commitMessageFile, []byte(commitMessage), 0644)
			if err != nil {
				fmt.Println(err)
				return
			}
			// print out that the commit message is written to the file
			fmt.Printf("Commit message written to %s\n", commitMessageFile)
			// print out usage instructions
			fmt.Printf("To use this commit message, run:\n\n\tgit commit -F %s\n\n", commitMessageFile)
		},
	}

	rootCmd.Flags().BoolVarP(&explain, "explain", "e", false, "Turn on console output for intermediate results")
	rootCmd.Flags().StringVarP(&hint, "hint", "", "", "Provide a hint for the commit message")
	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		return
	}
}
