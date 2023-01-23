package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	introMsg = "Is it Prime?\n------------\nEnter a whole number, and we'll tell you if it is prime number or not. Enter q to quit.\n"
)

func main() {
	intro()

	doneChan := make(chan bool)

	go getUserInput(doneChan, os.Stdin, os.Stdout)

	<-doneChan

	close(doneChan)

}

func intro() {
	fmt.Println(introMsg)
	prompt()
}

func prompt() {
	fmt.Print("-> ")
}

func isPrime(n int) (bool, string) {
	if n == 0 || n == 1 {
		return false, fmt.Sprintf("%d is not prime by defination!", n)
	}

	if n < 0 {
		return false, "Negative numbers are not prime by defination"
	}

	for i := 2; i <= n/2; i++ {
		if n%i == 0 {
			return false, fmt.Sprintf("%d is not prime by defination because it is divisible by %d", n, i)
		}
	}

	return true, fmt.Sprintf("%d is prime", n)
}

func getUserInput(c chan bool, reader io.Reader, writer io.Writer) {

	scanner := bufio.NewScanner(reader)

	for {

		res, done := checkNumbers(scanner)
		if done {
			c <- done
			return
		}

		writer.Write([]byte(res + "\n"))
		prompt()
	}
}

func checkNumbers(scanner *bufio.Scanner) (string, bool) {

	scanner.Scan()

	if strings.EqualFold(scanner.Text(), "q") {
		return "", true
	}

	if scanner.Text() == "" {
		return "please enter a whole number", true
	}

	n, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return "please enter a whole number", true
	}

	_, msg := isPrime(n)

	return msg, false
}
