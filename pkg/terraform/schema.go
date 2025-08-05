package terraform

import (
	_ "embed"
	"fmt"
)

//go:embed schema/schema.json
var schema string

func (m *Mixin) PrintSchema() {
	fmt.Fprint(m.Out, schema)
}
