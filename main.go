package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	ui "github.com/gizak/termui"
	l7g "github.com/immesys/chirp-l7g"
	"github.com/jacobsa/go-serial/serial"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <tty>\n", os.Args[0])
	}
	// Set up options.
	options := serial.OpenOptions{
		PortName:        os.Args[1],
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	// Make sure to close it later.
	defer port.Close()

	br := bufio.NewReader(port)
	initui()
	_ = br
	processLoop(br)
	for {
		time.Sleep(1 * time.Second)
	}
}

var plots []*ui.LineChart
var pars []*ui.Par
var tempbox *ui.Par

func initui() {
	if err := ui.Init(); err != nil {
		panic(err)
	}
	ui.Handle("/sys/kbd/q", func(ui.Event) {
		// press q to quit
		ui.StopLoop()
		os.Exit(0)
	})

	tempbox = ui.NewPar("External Temp Sensor: <UNK>")
	tempbox.Height = 3
	plots = make([]*ui.LineChart, 18)
    pars = make([]*ui.Par, 18)
	for src := 0; src < 3; src++ {
		for i := 0; i < 6; i++ {
			dispsrc := src
			if i < 3 {
				dispsrc += 3
			}
			lc := ui.NewLineChart()
			//lc.BorderLabel = "dot-mode Line Chart"
			lc.AxesColor = ui.ColorWhite
			lc.LineColor = ui.ColorRed | ui.AttrBold
			lc.Mode = "dot"
			lc.Data = []float64{1, 2, 3}
			lc.Height = 8
			//lc.BorderLabel = fmt.Sprintf("ASIC %d FROM %d", i+1, dispsrc+1)
			plots[src*6+i] = lc
            pars[src*6+i] = ui.NewPar("")
            pars[src*6+i].Height = 6
		}
	}
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, tempbox),
		),
		ui.NewRow(
			ui.NewCol(2, 0, plots[0]),
            ui.NewCol(2, 0, pars[0]),
			ui.NewCol(2, 0, plots[1]),
            ui.NewCol(2, 0, pars[1]),
            ui.NewCol(2, 0, plots[2]),
            ui.NewCol(2, 0, pars[2])),
		ui.NewRow(
			ui.NewCol(2, 0, plots[3]),
            ui.NewCol(2, 0, pars[3]),
			ui.NewCol(2, 0, plots[4]),
            ui.NewCol(2, 0, pars[4]),
            ui.NewCol(2, 0, plots[5]),
            ui.NewCol(2, 0, pars[5])),
	    ui.NewRow(
			ui.NewCol(2, 0, plots[0+6]),
            ui.NewCol(2, 0, pars[0+6]),
			ui.NewCol(2, 0, plots[1+6]),
            ui.NewCol(2, 0, pars[1+6]),
            ui.NewCol(2, 0, plots[2+6]),
            ui.NewCol(2, 0, pars[2+6])),
		ui.NewRow(
			ui.NewCol(2, 0, plots[3+6]),
            ui.NewCol(2, 0, pars[3+6]),
			ui.NewCol(2, 0, plots[4+6]),
            ui.NewCol(2, 0, pars[4+6]),
            ui.NewCol(2, 0, plots[5+6]),
            ui.NewCol(2, 0, pars[5+6])),
	    ui.NewRow(
			ui.NewCol(2, 0, plots[0+12]),
            ui.NewCol(2, 0, pars[0+12]),
			ui.NewCol(2, 0, plots[1+12]),
            ui.NewCol(2, 0, pars[1+12]),
            ui.NewCol(2, 0, plots[2+12]),
            ui.NewCol(2, 0, pars[2+12])),
		ui.NewRow(
			ui.NewCol(2, 0, plots[3+12]),
            ui.NewCol(2, 0, pars[3+12]),
			ui.NewCol(2, 0, plots[4+12]),
            ui.NewCol(2, 0, pars[4+12]),
            ui.NewCol(2, 0, plots[5+12]),
            ui.NewCol(2, 0, pars[5+12])))


	// calculate layout
	ui.Body.Align()

	ui.Render(ui.Body)
	go ui.Loop()

}
func processLoop(br *bufio.Reader) {

	buf := make([]byte, 82+64*3)
	for {
		//Scan for "cafebabe"
		for {
			c, err := br.ReadByte()
			if err != nil {
				panic(err)
			}
			if c != 'c' {
				continue
			}
			c, err = br.ReadByte()
			if err != nil {
				panic(err)
			}
			if c != 'a' {
				continue
			}
			c, err = br.ReadByte()
			if err != nil {
				panic(err)
			}
			if c != 'f' {
				continue
			}
			c, err = br.ReadByte()
			if err != nil {
				panic(err)
			}
			if c != 'e' {
				continue
			}
			c, err = br.ReadByte()
			if err != nil {
				panic(err)
			}
			if c != 'b' {
				continue
			}
			c, err = br.ReadByte()
			if err != nil {
				panic(err)
			}
			if c != 'a' {
				continue
			}
			c, err = br.ReadByte()
			if err != nil {
				panic(err)
			}
			if c != 'b' {
				continue
			}
			c, err = br.ReadByte()
			if err != nil {
				panic(err)
			}
			if c != 'e' {
				continue
			}
			break
		}
		//read payload
		_, err := io.ReadFull(br, buf)
		if err != nil {
			panic(err)
		}
		update(buf)
	}
}

var updateNum int
func update(buf []byte) {
    updateNum++
	//The first 76 bytes are the radio packet
	ch := l7g.ChirpHeader{}
	l7g.LoadChirpHeaderNoRecovery(1, buf[:82], &ch)
	if ch.Temperature != nil {
		tempbox.Text = fmt.Sprintf("External Temp Sensor: %.02f C", *ch.Temperature)
	} else {
		tempbox.Text = "External Temp Sensor: Not Present"
	}

	//The next 64*3 bytes are the raw IQs for the three recipients
	//lets just get mags
	mags := make([][]float64, 3)
    avg := make([]float64, 3)
    max := make([]float64, 3)
	for i := 0; i < 3; i++ {
		mags[i] = make([]float64, 16)
		dat := buf[82+(i*64):]
        mx := 0.0
        sm := 0.0
		for k := 0; k < 16; k++ {
			//IQ might be wrong order
			ival := int64(int16(binary.LittleEndian.Uint16(dat[k*4:])))
			qval := int64(int16(binary.LittleEndian.Uint16(dat[k*4+2:])))
			magsqr := ival*ival + qval*qval
			mag := math.Sqrt(float64(magsqr))
			mags[i][k] = mag
            if mag > mx {
                mx = mag
            }
            sm += mag
		}
        avg[i] = sm/16
        max[i] = mx
	}
	const numasics = 6
	if numasics == 4 {
		idx := 0
		for i := 0; i < 4; i++ {
			if i == int(ch.Primary) {
				continue
			}
			//sl[i].Data = mags[idx]
			idx++
		}
	} else if numasics == 6 {
		idx := 0
		if ch.Primary < 3 {
			for i := 3; i < 6; i++ {
				plotnum := int(ch.Primary*6) + i
                if plotnum > 18 {
                    panic(plotnum)
                }

				//fmt.Printf("for i=%d pri=%d plotnum=%d data=%v\n", i, ch.Primary, plotnum, mags[idx])
				plots[plotnum].Data = mags[idx]
				pars[plotnum].BorderLabel = fmt.Sprintf("ASIC %d FROM %d", i+1, ch.Primary+1)
                pars[plotnum].Text = fmt.Sprintf("Seq #: %d\nMax Index: %d\nMaximum: %.2f\nAverage: %.2f", updateNum, ch.MaxIndex[i], max[idx], avg[idx])
				idx++
			}
		} else {
			for i := 0; i < 3; i++ {
				plotnum := int(ch.Primary-3)*6 + i
                if plotnum > 18 {
                    panic(plotnum)
                }
				//fmt.Printf("for i=%d pri=%d plotnum=%d data=%v\n", i, ch.Primary, plotnum, mags[idx])
				//fmt.Printf("for i=%d pri=%d plotnum=%d\n", i, ch.Primary, plotnum)
				plots[plotnum].Data = mags[idx]
				pars[plotnum].BorderLabel = fmt.Sprintf("ASIC %d FROM %d", i+1, ch.Primary+1)
                pars[plotnum].Text = fmt.Sprintf("Seq #: %d\nMax Index: %d\nMaximum: %.2f\nAverage: %.2f", updateNum, ch.MaxIndex[i], max[idx], avg[idx])
				idx++
			}
		}
	}

	ui.Render(ui.Body)
}
