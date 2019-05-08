package common

import (
	"strings"
	"time"

	"github.com/rancher/mapper/convert"
	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	transitioningMap = map[string]string{}

	reverseErrorMap = map[string]bool{}

	errorMapping = map[string]bool{
		"Failed":      true,
		"Progressing": true,
	}

	progressMap = map[string]string{}

	doneMap = map[string]string{
		"Completed":   "activating",
		"Ready":       "unavailable",
		"Available":   "updating",
		"Progressing": "inactive",
	}
)

type State struct {
	State         string
	Error         bool
	Transitioning bool
	Message       string
}

type StateGetter interface {
	State() State
}

func StateFromConditionAndMeta(meta metav1.ObjectMeta, conditions []genericcondition.GenericCondition) (ret State) {
	state := ""
	error := false
	transitioning := false
	message := ""

	for _, c := range conditions {
		if (errorMapping[c.Type] && c.Status == "False") || c.Reason == "Error" {
			error = true
			message = c.Message
			break
		}
	}

	if !error {
		for _, c := range conditions {
			if reverseErrorMap[c.Type] && c.Status == "True" {
				error = true
				message = concat(message, c.Message)
			}
		}
	}

	for _, c := range conditions {
		newState, ok := transitioningMap[c.Type]
		if !ok {
			continue
		}

		if c.Status == "False" {
			error = true
			state = newState
			message = concat(message, c.Message)
		} else if c.Status == "Unknown" && state == "" {
			transitioning = true
			state = newState
			message = concat(message, c.Message)
		}
	}

	for _, c := range conditions {
		if state != "" {
			break
		}
		newState, ok := doneMap[c.Type]
		if !ok {
			continue
		}
		if c.Status == "False" {
			transitioning = true
			state = newState
			message = concat(message, c.Message)
		} else if c.Status == "Unknown" {
			error = true
			state = newState
			message = concat(message, c.Message)
		}
	}

	for _, c := range conditions {
		if state != "" {
			break
		}
		newState, ok := progressMap[c.Type]
		if !ok {
			continue
		}
		if c.Status == "True" {
			transitioning = true
			state = newState
			message = concat(message, c.Message)
		}
	}

	if state == "" {
		state = "active"
	}

	if error {
		error = true
	} else if transitioning {
		transitioning = true
	} else {
		transitioning = false
	}

	state = strings.ToLower(state)

	if meta.DeletionTimestamp != nil {
		state = "removing"
		transitioning = true

		finalizers := meta.Finalizers

		msg := message
		for _, cond := range conditions {
			if cond.Type == "Removed" && (cond.Status == "Unknown" || cond.Status == "False") && cond.Message != "" {
				msg = cond.Message
			}
		}

		if len(finalizers) > 0 {
			parts := strings.Split(finalizers[0], "wrangler.cattle.io/")
			f := parts[len(parts)-1]

			if f == "foregroundDeletion" {
				f = "object cleanup"
			}

			if len(msg) > 0 {
				msg = msg + "; waiting on " + f
			} else {
				msg = "waiting on " + f
			}
			message = msg
			if i, err := convert.ToTimestamp(meta.DeletionTimestamp); err == nil {
				if time.Unix(i/1000, 0).Add(5 * time.Minute).Before(time.Now()) {
					error = true
				}
			}
		}
	}

	ret.State = state
	ret.Error = error
	ret.Transitioning = transitioning
	ret.Message = message

	return ret
}

func concat(str, next string) string {
	if str == "" {
		return next
	}
	if next == "" {
		return str
	}
	if strings.EqualFold(str, next) {
		return str
	}
	return str + "; " + next
}
