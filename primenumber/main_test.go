package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func Test_IsPrime(t *testing.T) {

	cases := []struct {
		name         string
		n            int
		expectedBool bool
		expectedMsg  string
	}{
		{name: "zero", n: 0, expectedBool: false, expectedMsg: "0 is not prime by defination!"},
		{name: "one", n: 1, expectedBool: false, expectedMsg: "1 is not prime by defination!"},
		{name: "prime", n: 3, expectedBool: true, expectedMsg: "3 is prime"},
		{name: "not prime", n: 4, expectedBool: false, expectedMsg: "4 is not prime by defination because it is divisible by 2"},
		{name: "negative", n: -1, expectedBool: false, expectedMsg: "Negative numbers are not prime by defination"},
	}

	for _, tt := range cases {

		t.Run(tt.name, func(t *testing.T) {
			result, msg := isPrime(tt.n)

			if tt.expectedBool != result {
				t.Errorf("Expected '%v'; got '%v'", tt.expectedBool, result)
			}

			if tt.expectedMsg != msg {
				t.Errorf("expect '%v'; got '%v'", tt.expectedMsg, msg)
			}
		})

	}

}

func Test_prompt(t *testing.T) {

	oldStdOut := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = w

	prompt()

	_ = w.Close()

	os.Stdout = oldStdOut

	out, _ := io.ReadAll(r)

	log.Println(len(string(out)))

	if string(out) != "-> " {
		t.Errorf("Expect prompt to be '-> '; but got %v", string(out))
	}

}

func Test_intro(t *testing.T) {
	oldStdout := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = w
	intro()

	_ = w.Close()

	os.Stdout = oldStdout

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	result := string(out)
	expected := "Is it Prime"

	if !strings.Contains(result, expected) {
		t.Errorf("the intro message does not match; we expect '%v' but got '%v'", expected, result)
	}
}

func Test_checkNumbers(t *testing.T) {

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "prime", input: "7", expected: "7 is prime"},
		{name: "quit", input: "q", expected: ""},
		{name: "empty", input: "", expected: "please enter a whole number"},
		{name: "non-numeric", input: "sss", expected: `please enter a whole number`},
		{name: "decimal", input: "1.1", expected: `please enter a whole number`},
		{name: "one", input: "1", expected: "1 is not prime by defination!"},
		{name: "negative", input: "-1", expected: "Negative numbers are not prime by defination"},
	}

	for _, tt := range testCases {

		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			reader := bufio.NewScanner(input)

			res, _ := checkNumbers(reader)

			if res != tt.expected {
				t.Errorf("incorrect value retuened; got '%v'; expected '%v'", res, tt.expected)
			}
		})
	}
}

func Test_getUserInput(t *testing.T) {

	doneChan := make(chan bool)

	var stdin bytes.Buffer

	stdin.Write([]byte("1\nq\n"))

	go getUserInput(doneChan, &stdin, io.Discard)

	res := <-doneChan
	close(doneChan)

	if !res {
		t.Errorf("expected the program to end; got %v", res)
	}
}
