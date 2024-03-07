package dynamoconf

import (
	_ "embed"
)

//go:embed table.json
var Table []byte
