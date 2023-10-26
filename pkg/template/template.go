package template

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	t "html/template"

	"github.com/maxgio92/krawler/pkg/utils/matrix"
)

type MultiplexTemplate struct {
	templates []string
	vars      map[string][]string
}

type Option func(t *MultiplexTemplate)

func WithTemplates(templates ...string) Option {
	return func(t *MultiplexTemplate) {
		t.templates = templates
	}
}

func WithVariables(vars map[string][]string) Option {
	return func(t *MultiplexTemplate) {
		t.vars = vars
	}
}

func NewMultiplexTemplate(opts ...Option) *MultiplexTemplate {
	t := new(MultiplexTemplate)
	for _, f := range opts {
		f(t)
	}

	return t
}

// Run returns a list of strings of executed templates from a list of template string, by applying
// a set of variables, that can have multiple values.
// The expecetd options are:
//   - slice of template strings: the strings of the templates to execute.
//   - variables as map string to slice of strings. Each map's key represents the name of a variable,
//     that should match related variable reference in the template.
//     Each map's item should be a slice of strings for which each item represents a single variable's value.
//     Variables that are not referenced in the templates are ignored.
//
// The output is a slice of template strings, that contains all the combinations derived from the expasion
// of the variables.
func (t *MultiplexTemplate) Run() ([]string, error) {
	var res []string
	for _, v := range t.templates {
		v := v
		r, err := doRun(v, t.vars)
		if err != nil {
			return nil, err
		}
		res = append(res, r...)
	}

	return res, nil
}

func doRun(templateString string, vars map[string][]string) ([]string, error) {
	supportedVariables, err := GetSupportedVariables(templateString)
	if err != nil {
		return nil, err
	}

	// If the template string does not contain variables
	// return the template string directly.
	if len(supportedVariables) == 0 {
		return []string{templateString}, nil
	}

	// Populate the inventory.
	inventory := make(map[string][]string)
	for _, key := range supportedVariables {
		for _, v := range vars[key] {
			if v != "" {
				inventory[key] = append(inventory[key], v)
			}
		}
	}

	templateRegex, err := generateTemplateRegex(supportedVariables)
	if err != nil {
		return nil, err
	}
	templatePattern := regexp.MustCompile(templateRegex)

	ss, err := cutTemplateString(templateString, closeDelimiter)
	if err != nil {
		return nil, err
	}

	templateParts := []TemplatePart{}

	for _, s := range ss {

		// match are the template parts matched against the template regex.
		templatePartMatches := templatePattern.FindStringSubmatch(s)

		// name is the variable data structure to apply the template part to.
		for i, variableName := range templatePattern.SubexpNames() {

			// discard first variable name match and ensure a template part matched.
			if i > 0 && i <= len(templatePartMatches) && templatePartMatches[i] != "" {
				y := len(templateParts)

				templateParts = append(templateParts, TemplatePart{
					TemplateString:  templatePartMatches[i],
					MatchedVariable: variableName,
				})

				templateParts[y].Points = []string{}
				templateParts[y].TemplateString = strings.ReplaceAll(
					templateParts[y].TemplateString,
					openDelimiter+` `+cursor+variableName+` `+closeDelimiter,
					openDelimiter+` `+cursor+` `+closeDelimiter,
				)
				templateParts[y].Template = t.New(fmt.Sprintf("%d", y))
				templateParts[y].Template, err = templateParts[y].Template.Parse(templateParts[y].TemplateString)
				if err != nil {
					return nil, err
				}

				// for each item (variable name) of MatchedVariable
				// compose one Template and `execute()` it
				for _, value := range inventory[variableName] {
					o := new(bytes.Buffer)
					err = templateParts[y].Template.Execute(o, value)
					if err != nil {
						return nil, err
					}

					templateParts[y].Points = append(templateParts[y].Points.([]string), o.String())
				}
			}
		}
	}

	matrixColumns := []matrix.Column{}

	for _, part := range templateParts {
		matrixColumns = append(matrixColumns, part.Column)
	}

	if len(matrixColumns) <= 0 {
		return nil, fmt.Errorf("cannot multiplex template: the template contains syntax errors")
	}

	result, err := matrix.GetColumnOrderedCombinationRows(matrixColumns)
	if err != nil {
		return nil, err
	}

	return result, nil
}
