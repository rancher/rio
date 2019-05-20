package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rancher/axe/throwing"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rivo/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	showSystem    = false
	servicePrefix = &[]string{""}[0]
	podPrefix     = &[]string{""}[0]
)

func newSelectionList() *tview.List {
	list := tview.NewList()
	list.SetBackgroundColor(tcell.ColorGray)
	list.SetMainTextColor(tcell.ColorBlack)
	list.SetSecondaryTextColor(tcell.ColorPurple)
	list.SetShortcutColor(defaultBackGroundColor)
	return list
}

func newContainerSelectionList(t *throwing.TableView) *tview.List {
	list := newSelectionList()
	containers, err := listRioContainer(t.GetSelectionName(), t.GetClientSet())
	if err != nil {
		t.UpdateStatus(err.Error(), true)
		return nil
	}
	for _, c := range containers {
		list.AddItem(c, "", ' ', nil)
	}
	return list
}

func listRioContainer(name string, clientset *kubernetes.Clientset) ([]string, error) {
	namespace, name := kv.Split(name, "/")
	if name == "" {
		name = namespace
		namespace = "default"
	}
	name, version := kv.Split(name, ":")
	selector := fmt.Sprintf("app=%s", name)
	if version != "" {
		selector += fmt.Sprintf(",version=%s", version)
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}
	var containers []string
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			containers = append(containers, container.Name)
		}
		for _, container := range pod.Spec.InitContainers {
			containers = append(containers, container.Name)
		}
	}

	return containers, nil

}

/*
	general rio operation(inspect, edit, exec, logs, create)
*/
func inspect(format string, t *throwing.TableView) {
	name := t.GetSelectionName()
	outBuffer := &strings.Builder{}
	errBuffer := &strings.Builder{}
	args := []string{"inspect", "--format", format, "-t", t.GetResourceKind(), name}
	cmd := exec.Command("rio", args...)
	cmd.Stdout = outBuffer
	cmd.Stderr = errBuffer
	if err := cmd.Run(); err != nil {
		t.UpdateStatus(errBuffer.String(), true)
		return
	}

	inspectBox := tview.NewTextView()
	inspectBox.SetDynamicColors(true).SetBackgroundColor(defaultBackGroundColor)
	inspectBox.SetText(outBuffer.String())

	newpage := tview.NewPages().AddPage("inspect", inspectBox, true, true)
	t.SwitchPage(t.GetCurrentPage(), newpage)
}

func hit(t *throwing.TableView) {
	endpointCol := 0
	for i := 0; i < t.GetColumnCount(); i++ {
		if strings.Contains(t.GetCell(0, i).Text, "ENDPOINT") || strings.Contains(t.GetCell(0, i).Text, "URL") {
			endpointCol = i
			break
		}
	}
	row, _ := t.GetSelection()
	endpoint := strings.Trim(t.GetCell(row, endpointCol).Text, " ")
	if endpoint == "" {
		return
	}

	args := []string{"-c", "curl --connect-timeout 10 -s -k " + endpoint}
	b := &strings.Builder{}
	eb := &strings.Builder{}
	cmd := exec.Command("bash", args...)
	cmd.Stdout = b
	cmd.Stderr = eb
	if err := cmd.Run(); err != nil {
		t.UpdateStatus(eb.String(), true)
		return
	}

	hitBox := tview.NewTextView()
	hitBox.SetDynamicColors(true).SetBackgroundColor(defaultBackGroundColor)
	hitBox.SetText(b.String())

	newpage := tview.NewPages().AddPage("hit", hitBox, true, true)
	t.SwitchPage(t.GetCurrentPage(), newpage)
}

func promote(t *throwing.TableView) {
	name := t.GetSelectionName()
	args := []string{"promote", name}
	cmd := exec.Command("rio", args...)
	if err := cmd.Run(); err != nil {
		t.UpdateStatus(err.Error(), true)
		return
	}
}

func edit(t *throwing.TableView) {
	name := t.GetSelectionName()
	args := []string{"edit", "--raw", "-t", t.GetResourceKind()}
	args = append(args, name)
	cmd := exec.Command("rio", args...)
	errorBuffer := &strings.Builder{}
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, errorBuffer, os.Stdin

	t.GetApplication().Suspend(func() {
		clearScreen()
		if err := cmd.Run(); err != nil {
			t.UpdateStatus(errorBuffer.String(), true)
		}
		return
	})
}

func logs(container string, t *throwing.TableView) {
	name := t.GetSelectionName()
	args := []string{"logs", "-f"}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, name)
	if showSystem {
		args = append([]string{"--system"}, args...)
	}
	cmd := exec.Command("rio", args...)

	logbox := tview.NewTextView()
	{
		logbox.SetTitle(fmt.Sprintf("logs - (%s)", name))
		logbox.SetBorder(true)
		logbox.SetTitleColor(tcell.ColorPurple)
		logbox.SetDynamicColors(true)
		logbox.SetBackgroundColor(tcell.ColorBlack)
		logbox.SetChangedFunc(func() {
			logbox.ScrollToEnd()
			t.GetApplication().Draw()
		})
		logbox.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				cmd.Process.Kill()
			}
		})
	}

	cmd.Stdout = tview.ANSIWriter(logbox)
	go func() {
		if err := cmd.Run(); err != nil {
			return
		}
	}()

	newpage := tview.NewPages().AddPage("logs", logbox, true, true)
	t.SwitchPage(t.GetCurrentPage(), newpage)
}

func execute(container string, t *throwing.TableView) {
	name := t.GetSelectionName()
	shellArgs := []string{"/bin/sh", "-c", "TERM=xterm-256color; export TERM; [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c /bin/bash /dev/null || exec /bin/bash) || exec /bin/sh"}
	args := []string{"exec", "-it"}
	if container != "" {
		args = append(args, "-c", container)
	}
	if showSystem {
		args = append([]string{"--system"}, args...)
	}
	args = append(args, name)
	cmd := exec.Command("rio", append(args, shellArgs...)...)
	errorBuffer := &strings.Builder{}
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, errorBuffer, os.Stdin

	t.GetApplication().Suspend(func() {
		clearScreen()
		if err := cmd.Run(); err != nil {
			t.UpdateStatus(errorBuffer.String(), true)
			return
		}
		return
	})
}

func rm(t *throwing.TableView) {
	name := t.GetSelectionName()
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you want to delete %s %s?", t.GetResourceKind(), name)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				args := []string{"rm", "-t", t.GetResourceKind(), name}
				cmd := exec.Command("rio", args...)
				if err := cmd.Run(); err != nil {
					t.UpdateStatus(err.Error(), true)
					return
				}
				t.SwitchToRootPage()
			} else if buttonLabel == "Cancel" {
				t.SwitchToRootPage()
			}
		})
	t.InsertDialog("delete", t.GetCurrentPrimitive(), modal)
}

func viewPods(t *throwing.TableView) {
	name := t.GetSelectionName()
	podPrefix = &name
	t.GetTableView(podKind).RefreshManual()
	t.SwitchPage(podKind, t.GetTableView(podKind))
}

func revisions(t *throwing.TableView) {
	name := t.GetSelectionName()
	servicePrefix = &name
	t.GetTableView(serviceKind).RefreshManual()
	t.SwitchPage(serviceKind, t.GetTableView(serviceKind))
}

func escape(t *throwing.TableView) {
	kind := t.GetResourceKind()
	switch kind {
	case serviceKind:
		kind = appKind
	case podKind:
		kind = serviceKind
	}
	t.GetTableView(kind).RefreshManual()
	t.SwitchPage(kind, t.GetTableView(kind))
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
