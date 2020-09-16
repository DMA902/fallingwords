package game

import (
	"encoding/json"
	"strings"
)

// Word represents a single element of the JSON word list
type Word struct {
	Text string `json:"word"`
}

// LoadWordList loads all words from a given word list
//
// The word list must be able to be unmarshalled into a slice of Word
func LoadWordList(rawWordList []byte) (map[string]string, error) {
	words := []Word{}
	if err := json.Unmarshal(rawWordList, &words); err != nil {
		return nil, err
	}

	wordList := map[string]string{}
	for _, word := range words {
		upperWord := strings.ToUpper(word.Text)
		wordList[upperWord] = upperWord
	}

	return wordList, nil
}
