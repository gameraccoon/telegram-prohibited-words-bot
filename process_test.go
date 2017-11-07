package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWordsCountCalculation(t *testing.T) {
	assert := require.New(t)

	{
		testText := "Tested tests test testing"
		words := []string{"asd"}
		assert.Equal(0, len(findWords(testText, words)))
	}

	{
		testText := "Tested tests test testing"
		words := []string{"test"}
		assert.Equal(1, len(findWords(testText, words)))
	}

	{
		testText := "Tested tests test testing"
		words := []string{"Test"}
		assert.Equal(1, len(findWords(testText, words)))
	}

	{
		testText := "Tested tests test testing"
		words := []string{"Test", "Testing"}
		assert.Equal(2, len(findWords(testText, words)))
	}
}
