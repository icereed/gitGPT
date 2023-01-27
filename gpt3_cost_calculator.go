package main

import (
	gptTokenizer "github.com/wbrown/gpt_bpe"
)

func getNumTokens(sentence string) int {
	tokenizer := gptTokenizer.NewGPT2Encoder()
	return len(*tokenizer.Encode(&sentence))
}

func shortenToTokens(sentence string, tokens int) string {
	tokenizer := gptTokenizer.NewGPT2Encoder()
	encoded := *(tokenizer.Encode(&sentence))
	if len(encoded) > tokens {
		encoded = encoded[:tokens]
	}
	return tokenizer.Decode(&encoded)
}
