package main

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestWordsCountCalculation(t *testing.T) {
	assert := require.New(t)

	{
		testText := "Tested tests test testing"
		words := []string{strings.ToUpper("asd")}
		assert.Equal(0, len(findWords(testText, words)))
	}

	{
		testText := "Tested tests test testing"
		words := []string{strings.ToUpper("test")}
		assert.Equal(1, len(findWords(testText, words)))
	}

	{
		testText := "Tested tests test testing"
		words := []string{strings.ToUpper("Test")}
		assert.Equal(1, len(findWords(testText, words)))
	}

	{
		testText := "Tested tests test testing"
		words := []string{strings.ToUpper("Test"), strings.ToUpper("Testing")}
		assert.Equal(2, len(findWords(testText, words)))
	}
}
