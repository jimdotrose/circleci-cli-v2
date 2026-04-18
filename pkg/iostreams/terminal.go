package iostreams

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ReadPassword prompts the user for a secret value with input masked.
//
// When stdin is a terminal the input is hidden using golang.org/x/term.
// When stdin is not a terminal (piped input, tests) a plain line is read
// from In instead — this supports the `--with-token` flag pattern.
func (s *IOStreams) ReadPassword(prompt string) (string, error) {
	fmt.Fprint(s.ErrOut, prompt)

	if inFile, ok := s.In.(*os.File); ok && term.IsTerminal(int(inFile.Fd())) {
		pwd, err := term.ReadPassword(int(inFile.Fd()))
		fmt.Fprintln(s.ErrOut) // newline after masked input
		if err != nil {
			return "", err
		}
		return string(pwd), nil
	}

	// Non-terminal (piped input / test buffer): read a single line.
	scanner := bufio.NewScanner(s.In)
	scanner.Scan()
	fmt.Fprintln(s.ErrOut)
	return strings.TrimRight(scanner.Text(), "\r\n"), scanner.Err()
}
