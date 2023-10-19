package utils

import (
	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

// Based on the yq eval command. Simplified to just eval the expression in place.
// see https://github.com/mikefarah/yq/blob/master/cmd/evaluate_sequence_command.go#L138
// Note: the author does not guarantee the api will be consistent (see https://github.com/mikefarah/yq/issues/1486#issuecomment-1407781876)
func YqEval(expression, yaml string) (string, error) {

	yamlPref := yqlib.ConfiguredYamlPreferences
	// yamlPref.LeadingContentPreProcessing = false

	encoder := yqlib.NewYamlEncoder(2, false, yamlPref)
	decoder := yqlib.NewYamlDecoder(yamlPref)

	stringEvaluator := yqlib.NewStringEvaluator()
	return stringEvaluator.Evaluate(expression, yaml, encoder, decoder)
}

func YqEvalAll(expressions []string, yaml string) (string, error) {
	for _, expression := range expressions {
		var err error
		yaml, err = YqEval(expression, yaml)
		if err != nil {
			return "", err
		}
	}
	return yaml, nil
}
