package main

import (
	"fmt"

	"github.com/Machiel/slugify"
)

func main() {
	fmt.Println(slugify.Slugify("Hello, world!"))               // Will print: hello-world
	fmt.Println(slugify.Slugify("💻  I love this computer! 💻 ")) // Will print: i-love-this-computer

	dotSlugifier := slugify.New(slugify.Configuration{
		ReplaceCharacter: '.',
	})

	fmt.Println(dotSlugifier.Slugify("Hello, world!")) // Will print: hello.world

	numericOnlySlugifier := slugify.New(slugify.Configuration{
		IsValidCharacterChecker: func(c rune) bool {
			if c >= '0' && c <= '9' {
				return true
			}

			return false
		},
	})

	fmt.Println(numericOnlySlugifier.Slugify("3 eggs, 2 spoons of milk")) // Will print: 3-2

	replacementMapSlugifier := slugify.New(slugify.Configuration{
		ReplacementMap: map[rune]string{
			'a': "hello",
			'b': "hi",
		},
	})

	fmt.Println(replacementMapSlugifier.Slugify("a b")) // Will print: hello-hi
}
