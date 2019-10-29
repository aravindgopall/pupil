package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
)

// MODULE Main
// This tool is used to view logs with custom optimizations in terminal
//

// Pseudo code
// START
// Read the files
// Segregate the logs into three parts (INFO, WARN, ERROR)
// Map the keys 1, 2, 3 to the above buckets and based on the key
// pressed show the corresponding logs
// DONE

// This can be used to check if any errors occured after the new deployment

func main() {
	for _, arg := range os.Args[1:] {
		readFileAndProcess(arg)
	}
}

func readFileAndProcess(fn string) (err error) {

	file, err := os.Open(fn)
	defer file.Close()

	if err != nil {
		return err
	}

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	var line string
	var errors []string
	var infos []string
	var warns []string
	var allLogs []string

	for {
		line, err = reader.ReadString('\n')

		// Process the line here.
		allLogs = append(allLogs, line)

		if strings.Contains(line, "ERROR") {
			errors = append(allLogs, line)
		} else if strings.Contains(line, "INFO") {
			infos = append(infos, line)
		} else if strings.Contains(line, "WARN") {
			warns = append(warns, line)
		}

		if err != nil {
			break
		}
	}

	if err != io.EOF {
		fmt.Printf(" > Failed!: %v\n", err)
	}

	// Terminal setup
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	encoding.Register()

	if e = s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack))
	s.EnableMouse()
	s.Clear()

	quit := make(chan struct{})

	go pollEvents(s, quit, &allLogs, &errors, &infos, &warns)

	s.Show()

	go drawScreen(s, allLogs)

	<-quit
	s.Fini()

	return

}

func drawScreen(s tcell.Screen, logs []string) {
	for _, eachLog := range logs {
		io.WriteString(os.Stdout, eachLog)
	}
}

func pollEvents(s tcell.Screen, quit chan struct{}, allLogs *[]string, errors *[]string, infos *[]string, warns *[]string) {
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyEnter:
				close(quit)
				return
			case tcell.KeyRune:
				switch ev.Rune() {
				case '1':
					drawScreen(s, *allLogs)
				case '2':
					drawScreen(s, *errors)
				case '3':
					drawScreen(s, *infos)
				case '4':
					drawScreen(s, *warns)
				case 'q':
					close(quit)
					return
				}
				//s.Sync()
				//case tcell.KeyUp:
				//step := (vp.y0 - vp.y1) / 10
				//vp.y0 += step
				//vp.y1 += step
				//case tcell.KeyDown:
				//step := (vp.y0 - vp.y1) / 10
				//vp.y0 -= step
				//vp.y1 -= step
				//case tcell.KeyLeft:
				//step := (vp.x0 - vp.x1) / 10
				//vp.x0 += step
				//vp.x1 += step
				//case tcell.KeyRight:
				//step := (vp.x0 - vp.x1) / 10
				//vp.x0 -= step
				//vp.x1 -= step
			}
		//case *tcell.EventMouse:
		//x, y := ev.Position()
		//button := ev.Buttons()
		/*if button&tcell.WheelUp != 0 {
			bstr += " WheelUp"
		}*/
		// Only buttons, not wheel events
		//button &= tcell.ButtonMask(0xff)
		//switch ev.Buttons() {
		//case tcell.Button1:
		//zoom(s, 1, x, y)
		//case tcell.Button2:
		//zoom(s, 0, x, y)
		//}
		case *tcell.EventResize:
			s.Sync()
		}
	}
}

func limitLength(s string, length int) string {
	if len(s) < length {
		return s
	}

	return s[:length]
}
