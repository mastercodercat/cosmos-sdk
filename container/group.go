package container

import (
	"reflect"

	"github.com/pkg/errors"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

type groupResolver struct {
	typ          reflect.Type
	sliceType    reflect.Type
	idxsInValues []int
	providers    []*simpleProvider
	resolved     bool
	values       reflect.Value
}

type sliceGroupValueResolver struct {
	*groupResolver
}

func (g *sliceGroupValueResolver) resolve(c *container, _ Scope, caller containerreflect.Location) (reflect.Value, error) {
	// Log
	c.logf("Providing %v to %s from:", g.sliceType, caller.Name())
	c.indentLogger()
	for _, node := range g.providers {
		c.logf(node.ctr.Location.String())
	}
	c.dedentLogger()

	// Resolve
	if !g.resolved {
		res := reflect.MakeSlice(g.sliceType, 0, 0)
		for i, node := range g.providers {
			values, err := node.resolveValues(c)
			if err != nil {
				return reflect.Value{}, err
			}
			value := values[g.idxsInValues[i]]
			if value.Kind() == reflect.Slice {
				n := value.Len()
				for j := 0; j < n; j++ {
					res = reflect.Append(res, value.Index(j))
				}
			} else {
				res = reflect.Append(res, value)
			}
		}
		g.values = res
		g.resolved = true
	}

	return g.values, nil
}

func (g *groupResolver) resolve(_ *container, _ Scope, _ containerreflect.Location) (reflect.Value, error) {
	return reflect.Value{}, errors.Errorf("%v is an auto-group type and cannot be used as an input value, instead use %v", g.typ, g.sliceType)
}

func (g *groupResolver) addNode(n *simpleProvider, i int, c *container) error {
	g.providers = append(g.providers, n)
	g.idxsInValues = append(g.idxsInValues, i)

	constructorGraphNode, err := c.locationGraphNode(n.ctr.Location, n.scope)
	if err != nil {
		return err
	}

	typeGraphNode, err := c.typeGraphNode(g.sliceType)
	if err != nil {
		return err
	}

	c.addGraphEdge(constructorGraphNode, typeGraphNode)
	return nil
}