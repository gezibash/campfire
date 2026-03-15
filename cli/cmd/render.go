package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Breadcrumb represents a suggested next command for agentic consumers.
type Breadcrumb struct {
	Action      string `json:"action"`
	Cmd         string `json:"cmd"`
	Description string `json:"description"`
}

// Response wraps API results with metadata and breadcrumbs.
type Response struct {
	OK          bool         `json:"ok"`
	Data        interface{}  `json:"data"`
	Summary     string       `json:"summary"`
	Breadcrumbs []Breadcrumb `json:"breadcrumbs,omitempty"`
}

// BreadcrumbedItem embeds the original item fields alongside breadcrumbs.
type BreadcrumbedItem struct {
	Fields      map[string]interface{}
	Breadcrumbs []Breadcrumb
}

func (b BreadcrumbedItem) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, len(b.Fields)+1)
	for k, v := range b.Fields {
		m[k] = v
	}
	if len(b.Breadcrumbs) > 0 {
		m["breadcrumbs"] = b.Breadcrumbs
	}
	return json.Marshal(m)
}

// Column defines a single column for markdown table rendering.
type Column struct {
	Header string
	Field  string
}

// Columns is a slice of Column definitions.
type Columns []Column

// itemStr extracts a string value from a map, handling float64 coercion for JSON numbers.
func itemStr(item map[string]interface{}, key string) string {
	v, ok := item[key]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case json.Number:
		return val.String()
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// parseItems unmarshals a JSON array into a slice of maps.
func parseItems(raw []byte) ([]map[string]interface{}, error) {
	var items []map[string]interface{}
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// parseSingleItem unmarshals a JSON object into a map.
func parseSingleItem(raw []byte) (map[string]interface{}, error) {
	var item map[string]interface{}
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, err
	}
	return item, nil
}

// countItems returns the number of items in a JSON array.
func countItems(raw []byte) int {
	items, err := parseItems(raw)
	if err != nil {
		return 0
	}
	return len(items)
}

// outputList parses a JSON array, injects per-item breadcrumbs, wraps in an envelope, and prints.
func outputList(raw []byte, summary string, itemBreadcrumbsFn func(map[string]interface{}) []Breadcrumb, responseBreadcrumbsFn func([]map[string]interface{}) []Breadcrumb) {
	items, err := parseItems(raw)
	if err != nil {
		fmt.Println(string(raw))
		return
	}

	breadcrumbed := make([]BreadcrumbedItem, len(items))
	for i, item := range items {
		var breadcrumbs []Breadcrumb
		if itemBreadcrumbsFn != nil {
			breadcrumbs = itemBreadcrumbsFn(item)
		}
		breadcrumbed[i] = BreadcrumbedItem{Fields: item, Breadcrumbs: breadcrumbs}
	}

	var responseBreadcrumbs []Breadcrumb
	if responseBreadcrumbsFn != nil {
		responseBreadcrumbs = responseBreadcrumbsFn(items)
	}

	resp := Response{OK: true, Data: breadcrumbed, Summary: summary, Breadcrumbs: responseBreadcrumbs}
	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Println(string(raw))
		return
	}
	fmt.Println(string(out))
}

// outputSingle parses a JSON object, injects breadcrumbs, wraps in an envelope, and prints.
func outputSingle(raw []byte, summary string, itemBreadcrumbsFn func(map[string]interface{}) []Breadcrumb) {
	item, err := parseSingleItem(raw)
	if err != nil {
		fmt.Println(string(raw))
		return
	}

	var breadcrumbs []Breadcrumb
	if itemBreadcrumbsFn != nil {
		breadcrumbs = itemBreadcrumbsFn(item)
	}

	breadcrumbed := BreadcrumbedItem{Fields: item, Breadcrumbs: breadcrumbs}
	resp := Response{OK: true, Data: breadcrumbed, Summary: summary}
	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Println(string(raw))
		return
	}
	fmt.Println(string(out))
}

// escapeMDCell escapes pipe characters and newlines for markdown table cells.
func escapeMDCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// markdownList renders a JSON array as a GitHub-Flavored Markdown table with a summary header.
func markdownList(raw []byte, summary string, cols Columns) {
	items, err := parseItems(raw)
	if err != nil {
		fmt.Println(string(raw))
		return
	}

	fmt.Printf("**%s**\n\n", summary)

	if len(items) == 0 {
		fmt.Println("No results.")
		return
	}

	// Header row
	headers := make([]string, len(cols))
	separators := make([]string, len(cols))
	for i, col := range cols {
		headers[i] = col.Header
		separators[i] = "---"
	}
	fmt.Printf("| %s |\n", strings.Join(headers, " | "))
	fmt.Printf("| %s |\n", strings.Join(separators, " | "))

	// Data rows
	for _, item := range items {
		cells := make([]string, len(cols))
		for i, col := range cols {
			cells[i] = escapeMDCell(itemStr(item, col.Field))
		}
		fmt.Printf("| %s |\n", strings.Join(cells, " | "))
	}
}

// markdownDetail renders a JSON object as a markdown key-value list with a summary header.
func markdownDetail(raw []byte, summary string) {
	item, err := parseSingleItem(raw)
	if err != nil {
		fmt.Println(string(raw))
		return
	}

	fmt.Printf("**%s**\n\n", summary)

	for k, v := range item {
		fmt.Printf("- **%s**: %s\n", k, escapeMDCell(fmt.Sprintf("%v", v)))
	}
}

// markdownMutation prints a summary line for a mutation (create/update/delete).
func markdownMutation(summary string) {
	fmt.Printf("> %s\n", summary)
}
