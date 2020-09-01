package engine

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type ProviderAPI map[string]ProviderAPIResources

type ProviderAPIResources struct {
	Implicit map[string]Schema `json:"implicit"`
	Providers []Provider       `json:"providers"`
}

type Provider struct {
	Match      Schema
	Implicit   map[string]Schema
	Resource   string
	Action     string
	Parameters ProviderParameters
	Debug      bool
}

type ProviderParameters map[string]ProviderAPIResourcesParam

type ProviderAPIResourcesParam struct {
	Required bool
	Explicit Schema
	Implicit []ImplicitTask
}

type ImplicitTask struct {
	Name string `yaml:"resource"`
	Args map[string]interface{} `json:"args"`
	Type string
	Store  string
	Action string
}

func (t ImplicitTask) resolve(values map[string]interface{}) map[string]interface{} {
	var args = make(map[string]interface{})
	for paramName, paramValue := range t.Args {
		result := fmt.Sprint(paramValue)
		for resolvedParam, resolvedValue := range values {
			result = strings.Replace(
				result,
				fmt.Sprintf("$%v", resolvedParam),
				fmt.Sprint(resolvedValue),
				-1,
			)
		}
		args[paramName]	= result
	}
	return args
}

func (k ImplicitTask) getImplicitKeys() []string {
	re := regexp.MustCompile(`\$_[[:alpha:]]*`)
	var keys []string
	for _, value := range k.Args{
		for _, result := range re.FindAllString(fmt.Sprint(value), -1) {
			keys = append(keys, result)
		}
	}
	return keys
}

func (p Provider) toJsonSchemaDefs() Schema {
	argsSchema := make(Schema)
	for param, def := range p.Parameters {
		argsSchema[param] = def.toJsonSchema(p.Implicit, param)
	}
	return argsSchema
}

func (p Provider) toJsonSchema() Schema {
	//var required []string
	properties := make(map[string]interface{})
	argsSchema := make(Schema)
	/*for param, def := range p.Parameters {
		argsSchema[param] = def.toJsonSchema(p.Implicit, param)
		if def.Required {
			 required = append(required, param)
		}
	}*/
	for param, _ := range p.Parameters {
		argsSchema[param] = Schema{"$ref": fmt.Sprintf("%v",param)}
	}
	properties["Resource"] = Schema{
		"type": "object",
		"properties": Schema{
			"Name": Schema{
				"const": p.Resource,
			},
			"Type": Schema{
				"type": "string",
			},
			"args": Schema{
				"type": "object",
				"oeProperties": argsSchema,
			},
		},
		"additionalProperties": false,
	}
	properties["System"] = p.Match
	return Schema{
		//"$id": "provider.json",
		"title": "Provider",
		"type":     "object",
		"required": []string{"Resource", "System"},
		"properties": properties,
	}
}

func (p ProviderAPIResourcesParam) toJsonSchema(resourceImplicit map[string]Schema, name string) Schema {
	if len(p.Implicit) > 0 {
		implicitProperties := make(Schema)
		var implicitArgs []string
		for _, implicitTask := range p.Implicit {
			for _, arg := range implicitTask.Args {
				re := regexp.MustCompile(`\$_[[:alpha:]]*`)
				for _, match := range re.FindAll([]byte(fmt.Sprint(arg)), -1) {
					implicitArgs = append(implicitArgs, string(match[1:]))
				}
			}
		}
		var required []string
		for param, def := range resourceImplicit {
			if sort.SearchStrings(implicitArgs, param) != len(implicitArgs) {
				implicitProperties[param] = def
				required = append(required, param)
			}
		}
		implicit := Schema{
			"type": "object",
			"$anchor": "implicit",
			"properties": implicitProperties,
		}
		if len(required) > 0 {
			implicit["required"] = required
		}
		result := Schema{
			"$id": fmt.Sprintf("%v", name),
			"oneOf": []Schema{
				p.Explicit,
				implicit,
			},
		}
		if p.Required {
			result["oeRequired"] = true
		}
		return result
	}
	p.Explicit["$id"] = fmt.Sprintf("%v", name)
	if p.Required {
		p.Explicit["oeRequired"] = true
	}
	return p.Explicit
}

func (p Provider) getImplicitKeys() []string {
	var keys []string
	for key := range p.Implicit {
		keys = append(keys, key)
	}
	return keys
}

func (p ProviderAPIResourcesParam) getImplicitKeys() []string {
	var keys []string
	for _, task := range p.Implicit {
		keys = append(keys, task.getImplicitKeys()...)
	}
	return keys
}