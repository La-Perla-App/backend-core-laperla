package hooks

import (
	"encoding/json"
	"fmt"
)

func argsToString(args []any) string {
	argsJson, jsonErr := json.Marshal(args)

	argsStr := string(argsJson)

	// fallback to fmt.Sprintf %q if json marshalling fails
	if jsonErr != nil {
		var argsFmt []any
		for _, arg := range args {
			switch v := arg.(type) {
			case []byte:
				argsFmt = append(argsFmt, string(v))
			case float64, float32:
				argsFmt = append(argsFmt, fmt.Sprintf("%f", v))
			default:
				argsFmt = append(argsFmt, trimQueryContent(fmt.Sprintf("%s", arg)))
			}
		}
		argsStr = fmt.Sprintf("%q", argsFmt)
	}

	return argsStr
}
