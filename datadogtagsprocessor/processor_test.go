package datadogtagsprocessor

import (
	"path/filepath"
	"testing"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/plogtest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/ptracetest"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestProcessLogs_Merge_WithoutWildcards(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)

	oCfg.Mode = config.Merge
	oCfg.LogStatements = []config.ContextStatements{
		{
			Context: "log",
			Attributes: []string{
				"team",
			},
		},
		{
			Context: "resource",
			Attributes: []string{
				"k8s.deployment.name",
				"k8s.namespace.name",
				"k8s.pod.name",
				"k8s.container.name",
				"k8s.replicaset.name",
				"service.name",
				"service.version",
				"deployment.environment.name",
				"host.name",
			},
		},
	}
	sink := new(consumertest.LogsSink)
	p, err := factory.CreateLogs(t.Context(), processortest.NewNopSettings(metadata.Type), oCfg, sink)
	require.NoError(t, err)

	input, err := golden.ReadLogs(filepath.Join("testdata", "logs", "input.yaml"))
	require.NoError(t, err)
	expected, err := golden.ReadLogs(filepath.Join("testdata", "logs", "merge", "expected-without-wildcards.yaml"))
	require.NoError(t, err)

	require.NoError(t, p.ConsumeLogs(t.Context(), input))

	actual := sink.AllLogs()
	require.Len(t, actual, 1)

	require.NoError(t, plogtest.CompareLogs(expected, actual[0]))
}

func TestProcessLogs_Move_WithoutWildcards(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)

	oCfg.Mode = config.Move
	oCfg.LogStatements = []config.ContextStatements{
		{
			Context: "log",
			Attributes: []string{
				"team",
			},
		},
		{
			Context: "resource",
			Attributes: []string{
				"k8s.deployment.name",
				"k8s.namespace.name",
				"k8s.pod.name",
				"k8s.container.name",
				"k8s.replicaset.name",
				"service.name",
				"service.version",
				"deployment.environment.name",
				"host.name",
			},
		},
	}
	sink := new(consumertest.LogsSink)
	p, err := factory.CreateLogs(t.Context(), processortest.NewNopSettings(metadata.Type), oCfg, sink)
	require.NoError(t, err)

	input, err := golden.ReadLogs(filepath.Join("testdata", "logs", "input.yaml"))
	require.NoError(t, err)
	expected, err := golden.ReadLogs(filepath.Join("testdata", "logs", "move", "expected-without-wildcards.yaml"))
	require.NoError(t, err)

	require.NoError(t, p.ConsumeLogs(t.Context(), input))

	actual := sink.AllLogs()
	require.Len(t, actual, 1)

	require.NoError(t, plogtest.CompareLogs(expected, actual[0]))
}

func TestProcessTraces_Merge_WithoutWildcards(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)

	oCfg.Mode = config.Merge
	oCfg.TraceStatements = []config.ContextStatements{
		{
			Context: "span",
			Attributes: []string{
				"team",
			},
		},
		{
			Context: "resource",
			Attributes: []string{
				"k8s.deployment.name",
				"k8s.namespace.name",
				"k8s.pod.name",
				"k8s.container.name",
				"k8s.replicaset.name",
				"service.name",
				"service.version",
				"deployment.environment.name",
				"host.name",
			},
		},
	}
	sink := new(consumertest.TracesSink)
	p, err := factory.CreateTraces(t.Context(), processortest.NewNopSettings(metadata.Type), oCfg, sink)
	require.NoError(t, err)

	input, err := golden.ReadTraces(filepath.Join("testdata", "traces", "input.yaml"))
	require.NoError(t, err)
	expected, err := golden.ReadTraces(filepath.Join("testdata", "traces", "merge", "expected-without-wildcards.yaml"))
	require.NoError(t, err)

	require.NoError(t, p.ConsumeTraces(t.Context(), input))

	actual := sink.AllTraces()
	require.Len(t, actual, 1)

	require.NoError(t, ptracetest.CompareTraces(expected, actual[0]))
}

func TestProcessTraces_Move_WithoutWildcards(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)

	oCfg.Mode = config.Move
	oCfg.TraceStatements = []config.ContextStatements{
		{
			Context: "span",
			Attributes: []string{
				"team",
			},
		},
		{
			Context: "resource",
			Attributes: []string{
				"k8s.deployment.name",
				"k8s.namespace.name",
				"k8s.pod.name",
				"k8s.container.name",
				"k8s.replicaset.name",
				"service.name",
				"service.version",
				"deployment.environment.name",
				"host.name",
			},
		},
	}
	sink := new(consumertest.TracesSink)
	p, err := factory.CreateTraces(t.Context(), processortest.NewNopSettings(metadata.Type), oCfg, sink)
	require.NoError(t, err)

	input, err := golden.ReadTraces(filepath.Join("testdata", "traces", "input.yaml"))
	require.NoError(t, err)
	expected, err := golden.ReadTraces(filepath.Join("testdata", "traces", "move", "expected-without-wildcards.yaml"))
	require.NoError(t, err)

	require.NoError(t, p.ConsumeTraces(t.Context(), input))

	actual := sink.AllTraces()
	require.Len(t, actual, 1)

	require.NoError(t, ptracetest.CompareTraces(expected, actual[0]))
}
