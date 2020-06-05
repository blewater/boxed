package types

import (
	"fmt"
	"strconv"

	"github.com/alexeyco/simpletable"
)

const (
	ColorDefault = "\x1b[39m"

	ColorRed    = "\x1b[91m"
	ColorGreen  = "\x1b[32m"
	ColorBlue   = "\x1b[94m"
	ColorGray   = "\x1b[90m"
	ColorYellow = "\x1b[33m"
)

var style = simpletable.StyleRounded

func (workflow *Workflow) Display() {
	tree := simpletable.New()
	tree.SetStyle(style)

	fmt.Println()
	for idx, task := range workflow.Tasks {

		tree.Header = &simpletable.Header{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignCenter, Text: "Task " + strconv.Itoa(idx+1)},
				{Align: simpletable.AlignCenter, Text: "Workflow " + blue(workflow.Name)},
				{Align: simpletable.AlignCenter, Text: "Task Type"},
			},
		}

		tree.Body = &simpletable.Body{
			Cells: [][]*simpletable.Cell{
				{
					&simpletable.Cell{Text: "Name: " + blue(task.Name)},
					&simpletable.Cell{Text: status(task)},
					&simpletable.Cell{Text: taskType(task)},
				},
			},
		}

		tree.Footer = &simpletable.Footer{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignCenter, Span: 3, Text: red(task.Log)},
			},
		}

		tree.Println()
	}
}

func status(task *TaskType) string {
	if task.Completed && task.Log == "" {
		return green(fmt.Sprintf("Successful Run\n%s", task.CompletedAt.Format("Jan 02 15:04:05.000")))
	}

	if task.Completed {
		return red(fmt.Sprintf("Failed %s\n", task.CompletedAt.Format("Jan 02 15:04:05.000")))
	}

	return yellow("Pending...")
}

func taskType(task *TaskType) string {
	if task.IsServer {
		return blue("Server")
	}

	return blue("Remote")
}

func red(s string) string {
	return fmt.Sprintf("%s%s%s", ColorRed, s, ColorDefault)
}

func green(s string) string {
	return fmt.Sprintf("%s%s%s", ColorGreen, s, ColorDefault)
}

func blue(s string) string {
	return fmt.Sprintf("%s%s%s", ColorBlue, s, ColorDefault)
}

func yellow(s string) string {
	return fmt.Sprintf("%s%s%s", ColorYellow, s, ColorDefault)
}
