package questions

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/onsi/ginkgo/reporters/stenographer/support/go-isatty"
	"github.com/pkg/errors"
	"github.com/rancher/norman/parse/builder"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

type Questions struct {
	order  []*question
	result map[string]string
}

type question struct {
	q            v3.Question
	oldAnswer    string
	asked        bool
	inprogress   bool
	result       map[string]string
	forcePrompt  bool
	show         condition
	questions    map[string]*question
	subquestions []*question
}

func (q *question) ask() error {
	if q.inprogress {
		return fmt.Errorf("cycle detected in conditions asking %s", q.q.Variable)
	}

	q.inprogress = true
	defer func() {
		q.asked = true
		q.inprogress = false
	}()

	if q.asked {
		return nil
	}

	if q.oldAnswer != "" && !q.forcePrompt {
		q.result[q.q.Variable] = q.q.Default
		return nil
	}

	if ok, err := q.show.eval(); err != nil {
		return errors.Wrapf(err, "can not evaluate condition for %s", q.q.Variable)
	} else if ok {
		q.result[q.q.Variable], err = q.prompt()
		return err
	}

	if q.oldAnswer != "" {
		q.result[q.q.Variable] = q.oldAnswer
	} else {
		q.result[q.q.Variable] = q.q.Default
	}

	return nil
}

func PromptBool(text string, def bool) (bool, error) {
	msg := fmt.Sprintf("%s [y/N] ", text)
	defStr := "n"
	if def {
		msg = fmt.Sprintf("%s [Y/n] ", text)
		defStr = "y"
	}

	for {
		yn, err := Prompt(msg, defStr)
		if err != nil {
			return false, err
		}

		switch strings.ToLower(yn) {
		case "y":
			return true, nil
		case "n":
			return false, nil
		default:
			fmt.Println("Enter y or n")
		}
	}
}

func PrintToTerm(text ...interface{}) {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Print(text...)
	} else {
		fmt.Fprint(os.Stderr, text...)
	}
}

func PrintlnToTerm(text ...interface{}) {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Println(text...)
	} else {
		fmt.Fprintln(os.Stderr, text...)
	}
}

func PrintfToTerm(msg string, format ...interface{}) {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Printf(msg, format...)
	} else {
		fmt.Fprintf(os.Stderr, msg, format...)
	}
}

func Prompt(text, def string) (string, error) {
	for {
		PrintToTerm(text)
		answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}

		answer = strings.TrimSpace(answer)
		if answer == "" {
			answer = def
		}

		if answer == "" {
			continue
		}

		return answer, nil
	}
}

func (q *question) prompt() (string, error) {
	for {
		def := q.oldAnswer
		if def == "" {
			def = q.q.Default
		}

		choice := def
		if len(q.q.Options) > 0 {
			choice = strings.Join(q.q.Options, "/")
		}
		msg := fmt.Sprintf("[%s] %s [%s]: ", q.q.Variable, q.q.Description, choice)

		answer, err := Prompt(msg, def)
		if err != nil {
			return "", err
		}

		err = validate(answer, q.q)
		if err != nil {
			fmt.Printf("invalid value: %v\n", err)
			continue
		}

		return answer, nil
	}
}

func validate(val string, q v3.Question) error {
	field := &types.Field{}
	err := convert.ToObj(q, field)
	if err != nil {
		return err
	}

	if field.Type == "" {
		field.Type = "string"
	}

	converted, err := builder.ConvertSimple(field.Type, val, builder.Create)
	if err != nil {
		return err
	}

	return builder.CheckFieldCriteria(q.Variable, *field, converted)
}

type condition struct {
	result map[string]string
	checks [][2]string
	qs     map[string]*question
}

func newCondition(result map[string]string, qs map[string]*question, parentCondition, currentCondition string) condition {
	var checks [][2]string
	for _, check := range []string{parentCondition, currentCondition} {
		for _, part := range strings.Split(check, "&&") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			var vals [2]string
			vals[0], vals[1] = kv.Split(part, "=")
			checks = append(checks, vals)
		}
	}

	return condition{
		result: result,
		checks: checks,
		qs:     qs,
	}
}

func (c *condition) eval() (bool, error) {
	for _, check := range c.checks {
		q := c.qs[check[0]]
		if q == nil {
			continue
		}
		if err := q.ask(); err != nil {
			return false, err
		}
	}

	for _, check := range c.checks {
		if c.result[check[0]] != check[1] {
			return false, nil
		}
	}

	return true, nil
}

func NewQuestions(qs []v3.Question, answers map[string]string, forcePrompt bool) (*Questions, error) {
	result := map[string]string{}
	questions := map[string]*question{}
	var order []*question

	for _, q := range qs {
		qq := &question{
			q:           q,
			oldAnswer:   answers[q.Variable],
			result:      result,
			forcePrompt: forcePrompt,
			show:        newCondition(result, questions, "", q.ShowIf),
			questions:   questions,
		}

		order = append(order, qq)
		questions[q.Variable] = qq

		for _, subQ := range q.Subquestions {
			sq := v3.Question{}
			err := convert.ToObj(subQ, &q)
			if err != nil {
				return nil, err
			}

			qq := &question{
				q:           sq,
				oldAnswer:   answers[sq.Variable],
				result:      result,
				forcePrompt: forcePrompt,
				show:        newCondition(result, questions, q.ShowSubquestionIf, sq.ShowIf),
				questions:   questions,
			}

			questions[sq.Variable] = qq
		}
	}

	return &Questions{
		result: result,
		order:  order,
	}, nil
}

func (qs *Questions) Ask() (map[string]string, error) {
	for _, q := range qs.order {
		if err := q.ask(); err != nil {
			return nil, err
		}
	}

	return qs.result, nil
}
