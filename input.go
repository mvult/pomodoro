package main

import "github.com/eiannone/keyboard"

func readLine() (string, error) {
	var input string
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			return "", err
		}
		if key == keyboard.KeyEnter {
			break
		}
		input += string(char)
	}
	return input, nil
}
