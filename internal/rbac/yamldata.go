//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/rbac"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var (
	outputFile = flag.String("o", "", "output file")
	lang       = flag.String("lang", "go", "language to generate output for")
	kind       = flag.String("kind", "constants", "the kind of output to be generated")
)

type namespaceAction struct {
	varName string
	action  rtypes.NamespaceAction
}

// This generates the permission constants used on the frontend and backend for access control checks.
// The source of truth for RBAC is the `schema.yaml`, and this parses the YAML file, constructs the permission
// display names and saves the display names as constants.
func main() {
	flag.Parse()

	if *outputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	output, err := os.Create(*outputFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer output.Close()

	schema := rbac.RBACSchema

	var permissions = []types.Permission{}
	var namespaces = make([]rtypes.PermissionNamespace, len(schema.Namespaces))
	var actions = []namespaceAction{}
	for index, namespace := range schema.Namespaces {
		for _, action := range namespace.Actions {
			namespaces[index] = namespace.Name

			actionVarName := fmt.Sprintf("%s%sAction", sentencizeNamespace(namespace.Name.String()), toTitleCase(action.String()))
			actions = append(actions, namespaceAction{
				varName: actionVarName,
				action:  action,
			})

			permissions = append(permissions, types.Permission{
				Namespace: namespace.Name,
				Action:    action,
			})
		}
	}

	switch strings.ToLower(*lang) {
	case "go":
		if *kind == "constants" {
			generateGoConstants(output, permissions)
		} else if *kind == "namespace" {
			generateNamespaces(output, namespaces)
		} else if *kind == "action" {
			generateActions(output, actions)
		}
	case "ts":
		generateTSConstants(output, permissions)
	}
}

func generateTSConstants(output io.Writer, permissions []types.Permission) {
	fmt.Fprintln(output, "// Code generated by internal/rbac/yamldata. DO NOT EDIT.")
	for _, permission := range permissions {
		fmt.Fprintln(output)
		dn := permission.DisplayName()
		fmt.Fprintln(output, fmt.Sprintf("export const %sPermission = '%s'", sentencizeNamespace(dn), dn))
	}
}

func generateGoConstants(output io.Writer, permissions []types.Permission) {
	fmt.Fprintln(output, "// Code generated by yamldata. DO NOT EDIT.")
	fmt.Fprintln(output, "package rbac")
	for _, permission := range permissions {
		fmt.Fprintln(output)
		dn := permission.DisplayName()
		fmt.Fprintln(output, fmt.Sprintf("const %sPermission string = \"%s\"", sentencizeNamespace(dn), dn))
	}
}

func generateNamespaces(output io.Writer, namespaces []rtypes.PermissionNamespace) {
	fmt.Fprintln(output, "// Code generated by internal/rbac/yamldata. DO NOT EDIT.")
	fmt.Fprintln(output, "package types")
	fmt.Fprintln(output)

	var namespacesConstants = make([]string, len(namespaces))
	var namespaceVariableNames = make([]string, len(namespaces))
	for index, namespace := range namespaces {
		namespaceVarName := fmt.Sprintf("%sNamespace", sentencizeNamespace(namespace.String()))
		namespacesConstants[index] = fmt.Sprintf("const %s PermissionNamespace = \"%s\"", namespaceVarName, namespace)
		namespaceVariableNames[index] = namespaceVarName
	}

	fmt.Fprintf(output, rbacNamespaceTemplate, strings.Join(namespacesConstants, "\n"), strings.Join(namespaceVariableNames, ", "))
}

func generateActions(output io.Writer, namespaceActions []namespaceAction) {
	fmt.Fprintln(output, "// Code generated by internal/rbac/yamldata. DO NOT EDIT.")
	fmt.Fprintln(output, "package types")
	fmt.Fprintln(output)

	var namespaceActionConstants = make([]string, len(namespaceActions))
	for index, namespaceAction := range namespaceActions {
		namespaceActionConstants[index] = fmt.Sprintf("const %s NamespaceAction = \"%s\"", namespaceAction.varName, namespaceAction.action)
	}

	fmt.Fprintf(output, rbacActionTemplate, strings.Join(namespaceActionConstants, "\n"))
}

func sentencizeNamespace(permission string) string {
	separators := [2]string{"#", "_"}
	// Replace all separators with white spaces
	for _, sep := range separators {
		permission = strings.ReplaceAll(permission, sep, " ")
	}

	return toTitleCase(permission)
}

func toTitleCase(input string) string {
	words := strings.Fields(input)

	formattedWords := make([]string, len(words))

	for i, word := range words {
		formattedWords[i] = strings.Title(strings.ToLower(word))
	}

	return strings.Join(formattedWords, "")
}

const rbacNamespaceTemplate = `
// A PermissionNamespace represents a distinct context within which permission policies
// are defined and enforced.
type PermissionNamespace string

func (n PermissionNamespace) String() string {
	return string(n)
}

%s

// Valid checks if a namespace is valid and supported by Sourcegraph's RBAC system.
func (n PermissionNamespace) Valid() bool {
	switch n {
	case %s:
		return true
	default:
		return false
	}
}
`

const rbacActionTemplate = `
// NamespaceAction represents the action permitted in a namespace.
type NamespaceAction string

func (a NamespaceAction) String() string {
	return string(a)
}

%s
`
