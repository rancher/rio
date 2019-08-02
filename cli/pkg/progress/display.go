package progress

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	statusChar = "|/-\\"
)

type ConsoleWriter struct {
	index     int
	lastWidth int
}

func NewWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}
func modIndex(v int) int {
	v++
	return int(math.Mod(float64(v), 4.0))
}

func (c *ConsoleWriter) Display(format string, sleep int, args ...interface{}) {
	current, _ := fmt.Printf(fmt.Sprintf("\r%v %v", string(statusChar[c.index]), format), args...)
	if current < c.lastWidth {
		fmt.Printf(strings.Repeat(" ", c.lastWidth-current))
	}
	c.lastWidth = current
	c.index = modIndex(c.index)
	time.Sleep(time.Second * time.Duration(sleep))
}
