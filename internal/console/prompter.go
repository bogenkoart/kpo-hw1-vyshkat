package console

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrInputClosed — ввод закончился (EOF), продолжать диалог нельзя.
var ErrInputClosed = errors.New("ввод завершён")

// prompter — чтение ответов пользователя с повтором при некорректном вводе.
type prompter struct {
	in  *bufio.Scanner
	out io.Writer
}

func (p *prompter) line(question string) (string, error) {
	fmt.Fprint(p.out, question)
	if !p.in.Scan() {
		fmt.Fprintln(p.out)
		return "", ErrInputClosed
	}
	return strings.TrimSpace(p.in.Text()), nil
}

func (p *prompter) text(question string) (string, error) {
	for {
		s, err := p.line(question)
		if err != nil {
			return "", err
		}
		if s != "" {
			return s, nil
		}
		fmt.Fprintln(p.out, "  ! значение не может быть пустым")
	}
}

func (p *prompter) integer(question string, min, max int) (int, error) {
	for {
		s, err := p.line(question)
		if err != nil {
			return 0, err
		}
		n, convErr := strconv.Atoi(s)
		if convErr == nil && n >= min && n <= max {
			return n, nil
		}
		fmt.Fprintf(p.out, "  ! нужно целое число от %d до %d\n", min, max)
	}
}

func (p *prompter) number(question string) (float64, error) {
	for {
		s, err := p.line(question)
		if err != nil {
			return 0, err
		}
		f, convErr := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
		if convErr == nil && f >= 0 {
			return f, nil
		}
		fmt.Fprintln(p.out, "  ! нужно неотрицательное число, например 0.4")
	}
}
