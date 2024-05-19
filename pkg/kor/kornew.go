package kor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"strings"
)

type ResourceInfo struct {
	Name   string `json:"name"`
	Reason string `json:"reason,omitempty"`
}

func unusedResourceFormatter2(outputFormat string, outputBuffer bytes.Buffer, opts Opts, jsonResponse []byte) (string, error) {
	switch outputFormat {
	case "table":
		return outputBuffer.String(), nil
	case "json", "yaml":
		var resources map[string]map[string][]ResourceInfo
		if err := json.Unmarshal(jsonResponse, &resources); err != nil {
			return "", err
		}

		if !opts.PrintReason && outputFormat == "json" {
			// Create a map of namespaces with their corresponding maps of resource types and lists of resource names
			namespaces := make(map[string]map[string][]string)
			for namespace, resourceMap := range resources {
				for resourceType, infoSlice := range resourceMap {
					for _, info := range infoSlice {
						if _, ok := namespaces[namespace]; !ok {
							namespaces[namespace] = make(map[string][]string)
						}
						namespaces[namespace][resourceType] = append(namespaces[namespace][resourceType], info.Name)
					}
				}
			}
			// Marshal the map to JSON format
			modifiedJSONResponse, err := json.MarshalIndent(namespaces, "", "  ")
			if err != nil {
				return "", err
			}
			return string(modifiedJSONResponse), nil
		}

		// Marshal JSON response with reasons
		modifiedJSONResponse, err := json.MarshalIndent(resources, "", "  ")
		if err != nil {
			return "", err
		}
		return string(modifiedJSONResponse), nil
	default:
		return "", fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func FormatOutput2(resources map[string]map[string][]ResourceInfo, opts Opts) bytes.Buffer {
	var output bytes.Buffer
	switch opts.GroupBy {
	case "namespace":
		for namespace, diffs := range resources {
			output.WriteString(formatOutputForNamespace2(namespace, diffs, opts))
		}
	case "resource":
		for resource, diffs := range resources {
			output.WriteString(formatOutputForResource2(resource, diffs, opts))
		}
	}
	return output
}

func formatOutputForNamespace2(namespace string, resources map[string][]ResourceInfo, opts Opts) string {
	var buf strings.Builder
	table := tablewriter.NewWriter(&buf)
	table.SetHeader(getTableHeader(opts.GroupBy, opts.PrintReason))
	allEmpty := true
	var index int
	for resourceType, diff := range resources {
		for _, info := range diff {
			row := getTableRow(index, resourceType, info.Name)
			if opts.PrintReason && info.Reason != "" {
				row = append(row, info.Reason)
			}
			table.Append(row)
			allEmpty = false
			index++
		}
	}
	if allEmpty {
		if opts.Verbose {
			return fmt.Sprintf("No unused resources found in the namespace: %q\n", namespace)
		}
		return ""
	}

	table.Render()
	return fmt.Sprintf("Unused resources in namespace: %q\n%s\n", namespace, buf.String())
}

func formatOutputForResource2(resource string, resources map[string][]ResourceInfo, opts Opts) string {
	if len(resources) == 0 {
		if opts.Verbose {
			return fmt.Sprintf("No unused %ss found\n", resource)
		}
		return ""
	}
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader(getTableHeader(opts.GroupBy, opts.PrintReason))
	var index int
	for _, infos := range resources {
		for _, info := range infos {
			row := []string{info.Name}
			if opts.PrintReason && info.Reason != "" {
				row = append(row, info.Reason)
			}
			table.Append(row)
			index++
		}
	}
	table.Render()
	return fmt.Sprintf("Unused %ss:\n%s", resource, buf.String())
}

func appendResources2(resources map[string]map[string][]ResourceInfo, resourceType, namespace string, diff []ResourceInfo) {
	for _, d := range diff {
		if _, ok := resources[resourceType]; !ok {
			resources[resourceType] = make(map[string][]ResourceInfo)
		}
		resources[resourceType][namespace] = append(resources[resourceType][namespace], d)
	}
}