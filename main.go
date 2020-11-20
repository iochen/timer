package main

import (
	"fmt"
	"image/color"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

var lock bool

var dueTime time.Time

var fontSize = 320

func main() {
	version := fmt.Sprintf("%d", time.Now().UnixNano())
	if v := os.Getenv("VERSION"); v != "" {
		version = v
	}
	a := app.NewWithID("com.zincic.tool.timer" + "." + version)
	a.Settings().SetTheme(theme.LightTheme())
	a.SetIcon(resourceIconPng)

	w := a.NewWindow("Setting")
	w.Resize(fyne.NewSize(600, 500))
	w.SetContent(fyne.NewContainerWithLayout(layout.NewGridLayout(1), setAndStart()))

	w.SetCloseIntercept(a.Quit)

	w.ShowAndRun()
}

func setAndStart() fyne.CanvasObject {
	hour := widget.NewSelectEntry(makeList(23))
	hour.SetText(fmt.Sprintf("%02d", time.Now().Hour()+1))
	min := widget.NewSelectEntry(makeList(59))
	min.SetText("00")
	sec := widget.NewSelectEntry(makeList(59))
	sec.SetText("00")
	fs := widget.NewSelectEntry([]string{"100", "150", "200", "250", "300", "350", "400"})
	fs.SetText("320")

	form := &widget.Form{
		OnSubmit: func() {
			if lock {
				drv := fyne.CurrentApp().Driver()
				window := drv.CreateWindow("Error")
				window.SetContent(widget.NewLabel("Already opened!"))
				window.Show()
				return
			} else {
				var err error
				fontSize, err = strconv.Atoi(fs.Text)
				if err != nil {
					drv := fyne.CurrentApp().Driver()
					window := drv.CreateWindow("Error")
					window.SetContent(widget.NewLabel("font size not valid!"))
					window.Show()
					return
				}
				dueTime, err = time.Parse(time.RFC3339, fmt.Sprintf(time.Now().Format("2006-01-02T%s:%s:%sZ07:00"), hour.Text, min.Text, sec.Text))
				if err != nil {
					drv := fyne.CurrentApp().Driver()
					window := drv.CreateWindow("Error")
					window.SetContent(widget.NewLabel(err.Error()))
					window.Show()
					return
				}
				lock = true
				Show()
			}
		},
	}

	form.Append("Hour", hour)
	form.Append("Minute", min)
	form.Append("Second", sec)
	form.Append("Size", fs)

	query := widget.NewCard("Settings", "", form)
	return container.NewScroll(query)
}

func makeList(max int) []string {
	l := make([]string, max+1)
	for i := 0; i < max+1; i++ {
		d := strconv.Itoa(i)
		if len(d) == 1 {
			l[i] = "0" + d
		} else {
			l[i] = d
		}
	}
	return l
}

func Show() {
	drv := fyne.CurrentApp().Driver()
	if drv, ok := drv.(desktop.Driver); ok {
		w := drv.CreateSplashWindow()
		label := canvas.NewText("Please wait...", color.Black)
		label.TextSize = fontSize
		label.Alignment = fyne.TextAlignCenter
		go func() {
			belled := false
			tick := time.Tick(time.Second / 10)
			for range tick {
				d := dueTime.Sub(time.Now())
				label.Text = shortDur(d)
				if d.Minutes() < 15 {
					label.Color = color.NRGBA{255, 0, 0, 255}
				}
				if d.Seconds() > -1 && d.Seconds() <= 1 && !belled {
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "Time is up!",
						Content: fmt.Sprintf("It's %s now!", dueTime.Format("15:04:05")),
					})
					belled = true
				}
				label.Refresh()
			}
		}()
		button := widget.NewButton("exit", func() {
			w.Close()
			lock = false
		})
		ct := container.NewBorder(nil, container.NewCenter(button), nil, nil, label)
		w.SetContent(ct)
		w.SetFullScreen(true)
		w.CenterOnScreen()
		w.Show()
	}
}

func shortDur(d time.Duration) string {
	tpl := "%02d:%02d:%02d"
	var hour, min int
	t := d.Seconds()
	if t < 0 {
		t = -t
		t++ // cause int() a float value will convert 12.53 to 12
		tpl = "-" + tpl
	}

	for ; t > 3600; t -= 3600 {
		hour++
	}
	for ; t > 60; t -= 60 {
		min++
	}
	return fmt.Sprintf(tpl, hour, min, int(t))
}
