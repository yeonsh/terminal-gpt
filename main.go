package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"bytes"
	"os/exec"

	"github.com/otiai10/openaigo"
)


func main() {

	// Call openai api to get the command completion
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		fmt.Println("To use terminal-gpt please set the OPENAI_API_KEY environment variable")
		fmt.Println("You can get an API key from https://beta.openai.com/account/api-keys")
		fmt.Println("To set the environment variable, run:")
		fmt.Println("export OPENAI_API_KEY=<your key>")
		os.Exit(1)
	}

	withAliases := flag.Bool("with-aliases", false, "include aliases in the prompt")
	verbose := flag.Bool("verbose", false, "increase output verbosity")
    flag.Parse()

	question := strings.Join(flag.Args(), " ")

	keys := []string{"HOME", "USER", "SHELL"}
	environs := ""
	for _, key := range keys {
		value := os.Getenv(key)
		if value != "" {
			environs += fmt.Sprintf("%s=%s\n", key, value)
		}
	}

	shell := os.Getenv("SHELL")
	operatingSystem := runtime.GOOS
    if operatingSystem == "darwin" {
        operatingSystem = "macOS"
    }

	cwd, _ := os.Getwd()
	files := getFileList()

	prompt := fmt.Sprintf(`
You are an AI Terminal Copilot. Your job is to help users find the right terminal command in a %s on %s.

The user is asking for the following command:
'%s'

The user is currently in the following directory:
%s

That directory contains the following files:
%s

The user has several environment variables set, some of which are:
%s
`, shell, operatingSystem, question, cwd, files, environs)

    if *withAliases {
        prompt += fmt.Sprintf(`

        The user has the following aliases set:
%s`, getAliases())
    }

    if *verbose {
        fmt.Println(prompt)
    }

	client := openaigo.NewClient(openaiAPIKey)
	request := openaigo.CompletionRequestBody{
		Model:  "text-davinci-003",
		Prompt: []string{prompt},
        Temperature: 0.7,
        MaxTokens: 256,
        TopP: 1,
        Stop: []string{"`"},
        FrequencyPenalty: 0.0,
        PresencePenalty: 0.0,
	}
	response, err := client.Completion(nil, request)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    fmt.Println(response.Choices[0].Text)
}

func getFileList() string {
	dir, err := os.Open(".")

	if err != nil {
		// handle error
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		// handle error
	}

	fileList := make([]string, len(files))

	for i, file := range files {
		fileList[i] = file.Name()
	}

	return strings.Join(fileList, "\n")
}

func getAliases() string {
	// Use the `alias` command to list the defined aliases
	cmd := exec.Command("alias")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Parse the output of the `alias` command to extract the aliases
	aliases := make(map[string]string)
	for _, line := range strings.Split(out.String(), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "alias" {
			name := fields[1]
			value := strings.Join(fields[2:], " ")
			aliases[name] = value
		}
	}

    ret := ""
	for name, value := range aliases {
		ret += fmt.Sprintf("%s=%s\n", name, value)
	}

    return ret
}
