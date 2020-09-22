package util

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"os"
	"reflect"
)

type PrintStyle string

const (
	PrintStyleDefault  PrintStyle = "default"
	PrintStyleCsv      PrintStyle = "csv"
	PrintStyleHtml     PrintStyle = "html"
	PrintStyleMarkdown PrintStyle = "markdown"
	PrintStyleJson     PrintStyle = "json"
)

type PrintRequest struct {
	Rows       []table.Row
	Header     table.Row
	Data       interface{}
	PrintStyle PrintStyle
}

func Print(request PrintRequest) {
	if request.PrintStyle == PrintStyleJson {
		printJSON(request.Data)
		return
	}
	if request.Header != nil && request.Rows != nil {
		printTable(request.Rows, request.Header, request.PrintStyle)
	} else {
		printItem(request.Data, request.PrintStyle)
	}
}

func printJSON(any interface{}) {
	transformer := text.NewJSONTransformer("", "    ")
	fmt.Printf("%s", transformer(any))
}

func printTable(rows []table.Row, header table.Row, printType PrintStyle) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.Style().Format.Header = text.FormatTitle
	t.Style().Format.Footer = text.FormatDefault
	t.Style().Color.Header = text.Colors{text.Bold}
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.AppendFooter(table.Row{"Total:", len(rows)})
	if printType == PrintStyleDefault {
		t.Render()
	} else if printType == PrintStyleCsv {
		t.RenderCSV()
	} else if printType == PrintStyleHtml {
		t.RenderHTML()
	} else if printType == PrintStyleMarkdown {
		t.RenderMarkdown()
	}
}

func printItem(data interface{}, printStyle PrintStyle) {
	if reflect.TypeOf(data) == reflect.TypeOf(models.Cluster{}) {
		printCluster(data.(models.Cluster), printStyle)
	} else {
		color.Red("Not Implemented")
		os.Exit(1)
	}
}
