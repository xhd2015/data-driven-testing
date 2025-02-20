package goresolve

import "fmt"

func (vars Vars) ResolveRefs() error {
	mappingByNames := make(map[string]*Var, len(vars))
	for _, v := range vars {
		if v.Name == "_" {
			continue
		}
		mappingByNames[v.Name] = v
	}

	var traverse func(v *Def) error

	traverse = func(v *Def) error {
		refVarName := v.RefVarName
		if refVarName != "" {
			refVar := mappingByNames[refVarName]
			if refVar == nil {
				return fmt.Errorf("%s not found", refVarName)
			}
			refVar.HasRef = true
			v.RefVar = refVar
		}
		for _, child := range v.Children {
			err := traverse(child)
			if err != nil {
				return err
			}
		}
		return nil

	}
	for _, v := range vars {
		err := traverse(v.Def)
		if err != nil {
			return fmt.Errorf("%s: %w", v.Name, err)
		}
	}
	return nil
}
