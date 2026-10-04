package export

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// JSONExporter exports ExportData as JSON bytes.
type JSONExporter struct {
	Pretty bool
}

func (e *JSONExporter) Export(data ExportData) ([]byte, error) {
	// If ExportData contains "statistics", it will be serialized as part of the JSON object.
	if e.Pretty {
		return json.MarshalIndent(data, "", "  ")
	}
	return json.Marshal(data)
}

// CSVExporter exports ExportData as CSV bytes (flattening basic lists/rows).
type CSVExporter struct{}

func (e *CSVExporter) Export(data ExportData) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	var headers []string
	for k := range data {
		headers = append(headers, k)
	}

	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	var row []string
	for _, h := range headers {
		val := data[h]
		row = append(row, fmt.Sprintf("%v", val))
	}

	if err := writer.Write(row); err != nil {
		return nil, err
	}

	writer.Flush()
	return []byte(buf.String()), nil
}

// MarkdownExporter exports ExportData as a Markdown document.
type MarkdownExporter struct{}

func formatTitle(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func (e *MarkdownExporter) Export(data ExportData) ([]byte, error) {
	var buf strings.Builder
	buf.WriteString("# Exported Resource\n\n")

	for k, v := range data {
		buf.WriteString(fmt.Sprintf("## %s\n", formatTitle(k)))
		buf.WriteString(fmt.Sprintf("```json\n%v\n```\n\n", v))
	}

	return []byte(buf.String()), nil
}

// ExcelExporter exports ExportData as spreadsheet format.
type ExcelExporter struct{}

func (e *ExcelExporter) Export(data ExportData) ([]byte, error) {
	csvExp := &CSVExporter{}
	return csvExp.Export(data)
}

// YAMLExporter exports ExportData as YAML.
type YAMLExporter struct{}

func (e *YAMLExporter) Export(data ExportData) ([]byte, error) {
	return yaml.Marshal(data)
}

// XMLExporter exports ExportData as XML.
type XMLExporter struct{}

type xmlWrapper struct {
	XMLName xml.Name  `xml:"Export"`
	Items   []xmlItem `xml:"Item"`
}

type xmlItem struct {
	Key   string `xml:"Key,attr"`
	Value string `xml:",chardata"`
}

func (e *XMLExporter) Export(data ExportData) ([]byte, error) {
	var wrapper xmlWrapper
	wrapper.XMLName.Local = "Export"

	for k, v := range data {
		wrapper.Items = append(wrapper.Items, xmlItem{
			Key:   k,
			Value: fmt.Sprintf("%v", v),
		})
	}

	return xml.MarshalIndent(wrapper, "", "  ")
}
