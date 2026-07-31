package nats

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestManagedStreamDefinitionsIncludeGraphTask(t *testing.T) {
	streams := managedStreamDefinitions()

	var taskSubjects []string
	for _, stream := range streams {
		if stream.name == "task" {
			taskSubjects = stream.subjects
			break
		}
	}

	require.Contains(t, taskSubjects, "apps.panda-wiki.vector.task")
	require.Contains(t, taskSubjects, "apps.panda-wiki.graph.task")
}
