package processor

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var EmptyFilenameError = errors.New("filename can't be empty")
var EmptyLabelError = errors.New("label can't be empty")
var MissingInputFileError = errors.New("missing input file")

var logger = log.New(os.Stdout, "ImageProcessor: ", 0)

var xmlHeader = []byte("<?xml version=\"1.0\"?>\n")

// CommandChan receives commands for ImageProcessor
type CommandChan chan Command

type Processor interface {
	ProcessImage(filename string, width, height int, label string, top, left, right, bottom int) (result interface{}, err error)
}

// processorImpl is service that takes commands for processing images and their mapping
type processorImpl struct {
	unlabeledPath, labeledPath string
	inpChan                    CommandChan
}

// Command to execute on processorImpl
type Command struct {
	Filename string

	// width and height of original image
	Width, Height int

	// bounding box params
	Label         string
	Left, Top     int
	Right, Bottom int

	Resp chan interface{}
}

type pascalvoc struct {
	XMLName  xml.Name `xml:"annotation"`
	Folder   string   `xml:"folder"`
	Filename string   `xml:"filename"`
	Path     string   `xml:"path"`
	Database string   `xml:"source>database"`

	Width  int `xml:"size>width"`
	Height int `xml:"size>height"`
	Depth  int `xml:"size>depth"`

	Segmented int `xml:"segmented"`

	Name      string `xml:"object>name"`
	Pose      string `xml:"object>pose"`
	Truncated int    `xml:"object>truncated"`
	Difficult int    `xml:"object>difficult"`
	Xmin      int    `xml:"object>bndbox>xmin"`
	Ymin      int    `xml:"object>bndbox>ymin"`
	Xmax      int    `xml:"object>bndbox>xmax"`
	Ymax      int    `xml:"object>bndbox>ymax"`
}

func (p *processorImpl) ProcessImage(filename string, width, height int, label string, left, top, right, bottom int) (result interface{}, err error) {
	c := Command{
		Filename: filename,
		Width:    width,
		Height:   height,

		Label:  label,
		Left:   left,
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Resp:   make(chan interface{}, 1),
	}
	select {
	case p.inpChan <- c:
		select {
		case v, ok := <-c.Resp:
			if !ok {
				return nil, errors.New("reponse chan was closed")
			}

			switch r := v.(type) {
			case error:
				return nil, r
			default:
				return v, nil
			}
		case <-time.After(time.Second):
			return nil, errors.New("read response timeout exceeded")
		}
	case <-time.After(time.Second):
		return nil, errors.New("write command timeout exceeded")
	}
}

func (p *processorImpl) start() {
	go func() {
		for {
			select {
			case command, ok := <-p.inpChan:
				if !ok {
					logger.Fatalf("input channel is closed")
				}

				logger.Printf("received command %v\n", command)

				p.processCommand(command)
			}
		}
	}()
}

func (p *processorImpl) processCommand(c Command) {
	if c.Filename == "" {
		p.WriteResponse(c, nil, EmptyFilenameError)
		return
	}

	oldFilePath := path.Join(p.unlabeledPath, c.Filename)

	if _, err := os.Stat(oldFilePath); os.IsNotExist(err) {
		p.WriteResponse(c, nil, fmt.Errorf("%w (%s)", MissingInputFileError, oldFilePath))
		return
	}

	if strings.Trim(c.Label, " \n") == "" {
		p.WriteResponse(c, nil, EmptyLabelError)
		return;
	}

	doc := &pascalvoc{
		Folder:   filepath.Base(p.unlabeledPath),
		Filename: c.Filename,
		Path:     oldFilePath,
		Database: "Unknown",

		// Size
		Width:  c.Width,
		Height: c.Height,
		Depth:  3,

		Segmented: 0,

		Name:      c.Label,
		Pose:      "Unspecified",
		Truncated: 0,
		Difficult: 0,

		Xmin: c.Left,
		Ymin: c.Top,
		Xmax: c.Right,
		Ymax: c.Bottom,
	}

	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		p.WriteResponse(c, nil, fmt.Errorf("error marshalling document: %w", err))
		return
	}

	//logger.Println(string(output))

	newFilePath := path.Join(p.labeledPath, c.Filename)
	cleanName := strings.TrimSuffix(c.Filename, filepath.Ext(c.Filename))
	xmlPath := path.Join(p.labeledPath, cleanName+".xml")

	//logger.Printf("Writing xml to %s\n", xmlPath)

	file, err := os.Create(xmlPath)
	if err != nil {
		p.WriteResponse(c, nil, fmt.Errorf("error creating xml file: %w", err))
		return
	}

	written, err := file.Write(xmlHeader)
	if err != nil {
		p.WriteResponse(c, nil, fmt.Errorf("error writing to file: %w", err))
		return
	} else if written < len(xmlHeader) {
		p.WriteResponse(c, nil,fmt.Errorf("couldn't write all data: written %d instead of %d", written, len(output)))
		return
	}

	written, err = file.Write(output)
	if err != nil {
		p.WriteResponse(c, nil, fmt.Errorf("error writing to file: %w", err))
		return
	} else if written < len(output) {
		p.WriteResponse(c, nil,fmt.Errorf("couldn't write all data: written %d instead of %d", written, len(output)))
		return
	}

	if err := file.Close(); err != nil {
		p.WriteResponse(c, nil, fmt.Errorf("error flushing file: %w", err))
		return
	}

	if err = os.Rename(oldFilePath, newFilePath); err != nil {
		p.WriteResponse(c, nil, fmt.Errorf("error moving image: %w", err))
		return
	}

	p.WriteResponse(c, true, nil)
}

func (p *processorImpl) WriteResponse(c Command, result interface{}, err error) {
	if err != nil {
		logger.Printf("error processing command: %v\n", err)
		c.Resp <- err
	}

	select {
	case c.Resp <- result:
		// it's ok
	case <- time.After(time.Second):
		logger.Println("write response timeout exceeded")
	}
}

// NewImageProcessor creates new ImageProcessor, starts it and returns it
func NewImageProcessor(unlabeledPath, labeledPath string) (Processor, error) {
	if _, err := os.Stat(unlabeledPath); os.IsNotExist(err) {
		return nil, errors.New("unlabeled path doesn't exists")
	}

	if _, err := os.Stat(labeledPath); os.IsNotExist(err) {
		return nil, errors.New("labeled path doesn't exists")
	}

	p := &processorImpl{
		unlabeledPath: unlabeledPath,
		labeledPath:   labeledPath,
		inpChan:       make(CommandChan),
	}

	p.start()

	return p, nil
}
