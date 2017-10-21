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
		assert.Equal(0, calcWordsCount(strings.ToUpper(testText), words))
	}

	{
		testText := "Tested tests test testing"
		words := []string{strings.ToUpper("test")}
		assert.Equal(1, calcWordsCount(strings.ToUpper(testText), words))
	}

	{
		testText := "Tested tests test testing"
		words := []string{strings.ToUpper("Test")}
		assert.Equal(1, calcWordsCount(strings.ToUpper(testText), words))
	}

	{
		testText := "Tested tests test testing"
		words := []string{strings.ToUpper("Test"), strings.ToUpper("Testing")}
		assert.Equal(2, calcWordsCount(strings.ToUpper(testText), words))
	}
}
