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
		words := []string{"asd"}
		assert.Equal(0, calcWordsCount(strings.ToUpper(testText), words))
	}

	{
		testText := "Tested tests test testing"
		words := []string{"test"}
		assert.Equal(1, calcWordsCount(strings.ToUpper(testText), words))
	}

	{
		testText := "Tested tests test testing"
		words := []string{"Test"}
		assert.Equal(1, calcWordsCount(strings.ToUpper(testText), words))
	}

	{
		testText := "Tested tests test testing"
		words := []string{"Test", "Testing"}
		assert.Equal(2, calcWordsCount(strings.ToUpper(testText), words))
	}
}
