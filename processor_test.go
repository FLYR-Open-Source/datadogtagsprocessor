package datadogtagsprocessor

import (
	"path/filepath"
	"testing"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/plogtest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/ptracetest"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processortest"
)

// Test Logs Processing
func TestProcessLogs_Merge_WithoutWildcards(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)

	oCfg.LogStatements = []config.ContextStatements{
		{
			Mode:    config.Merge,
			Context: "log",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Merge,
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

func TestProcessLogs_Merge_WithWildcards(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)

	oCfg.LogStatements = []config.ContextStatements{
		{
			Mode:    config.Merge,
			Context: "log",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Merge,
			Context: "resource",
			Attributes: []string{
				"k8s.*",
				"service.*",
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
	expected, err := golden.ReadLogs(filepath.Join("testdata", "logs", "merge", "expected-with-wildcards.yaml"))
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

	oCfg.LogStatements = []config.ContextStatements{
		{
			Mode:    config.Move,
			Context: "log",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Move,
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

func TestProcessLogs_Move_WithWildcards(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)

	oCfg.LogStatements = []config.ContextStatements{
		{
			Mode:    config.Move,
			Context: "log",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Move,
			Context: "resource",
			Attributes: []string{
				"k8s.*",
				"service.*",
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
	expected, err := golden.ReadLogs(filepath.Join("testdata", "logs", "move", "expected-with-wildcards.yaml"))
	require.NoError(t, err)

	require.NoError(t, p.ConsumeLogs(t.Context(), input))

	actual := sink.AllLogs()
	require.Len(t, actual, 1)

	require.NoError(t, plogtest.CompareLogs(expected, actual[0]))
}

// Test Trace Processing
func TestProcessTraces_Merge_WithoutWildcards(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)

	oCfg.TraceStatements = []config.ContextStatements{
		{
			Mode:    config.Merge,
			Context: "span",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Merge,
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

func TestProcessTraces_Merge_WithWildcards(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)

	oCfg.TraceStatements = []config.ContextStatements{
		{
			Mode:    config.Merge,
			Context: "span",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Merge,
			Context: "resource",
			Attributes: []string{
				"k8s.*",
				"service.*",
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
	expected, err := golden.ReadTraces(filepath.Join("testdata", "traces", "merge", "expected-with-wildcards.yaml"))
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

	oCfg.TraceStatements = []config.ContextStatements{
		{
			Mode:    config.Move,
			Context: "span",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Move,
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

func TestProcessTraces_Move_WithWildcards(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)

	oCfg.TraceStatements = []config.ContextStatements{
		{
			Mode:    config.Move,
			Context: "span",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Move,
			Context: "resource",
			Attributes: []string{
				"k8s.*",
				"service.*",
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
	expected, err := golden.ReadTraces(filepath.Join("testdata", "traces", "move", "expected-with-wildcards.yaml"))
	require.NoError(t, err)

	require.NoError(t, p.ConsumeTraces(t.Context(), input))

	actual := sink.AllTraces()
	require.Len(t, actual, 1)

	require.NoError(t, ptracetest.CompareTraces(expected, actual[0]))
}

// Log Processing Benchmarks

// The processor mutates data in place (Move strips the matched attributes,
// Merge appends to ddtags), so each iteration must consume a fresh copy of
// the input. Reusing one object would make every iteration after the first
// process already-consumed data. The copy cost is included in each
// measurement. BenchmarkLogs_Baseline_Copy isolates it so it can be
// subtracted.
func benchmarkLogs(b *testing.B, statements []config.ContextStatements) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	oCfg.LogStatements = statements

	p, err := factory.CreateLogs(b.Context(), processortest.NewNopSettings(metadata.Type), oCfg, consumertest.NewNop())
	require.NoError(b, err)

	input, err := golden.ReadLogs(filepath.Join("testdata", "logs", "input.yaml"))
	require.NoError(b, err)

	for b.Loop() {
		cp := plog.NewLogs()
		input.CopyTo(cp)
		require.NoError(b, p.ConsumeLogs(b.Context(), cp))
	}
}

func BenchmarkLogs_Baseline_Copy(b *testing.B) {
	input, err := golden.ReadLogs(filepath.Join("testdata", "logs", "input.yaml"))
	require.NoError(b, err)

	for b.Loop() {
		cp := plog.NewLogs()
		input.CopyTo(cp)
	}
}

func BenchmarkLogs_Merge_WithoutWildcards(b *testing.B) {
	benchmarkLogs(b, []config.ContextStatements{
		{
			Mode:    config.Merge,
			Context: "log",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Merge,
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
	})
}

func BenchmarkLogs_Merge_WithWildcards(b *testing.B) {
	benchmarkLogs(b, []config.ContextStatements{
		{
			Mode:    config.Merge,
			Context: "log",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Merge,
			Context: "resource",
			Attributes: []string{
				"k8s.*",
				"service.*",
				"deployment.environment.name",
				"host.name",
			},
		},
	})
}

func BenchmarkLogs_Move_WithoutWildcards(b *testing.B) {
	benchmarkLogs(b, []config.ContextStatements{
		{
			Mode:    config.Move,
			Context: "log",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Move,
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
	})
}

func BenchmarkLogs_Move_WithWildcards(b *testing.B) {
	benchmarkLogs(b, []config.ContextStatements{
		{
			Mode:    config.Move,
			Context: "log",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Move,
			Context: "resource",
			Attributes: []string{
				"k8s.*",
				"service.*",
				"deployment.environment.name",
				"host.name",
			},
		},
	})
}

// Trace Processing Benchmarks

// See benchmarkLogs for why each iteration consumes a fresh copy and how
// BenchmarkTraces_Baseline_Copy fits in.
func benchmarkTraces(b *testing.B, statements []config.ContextStatements) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	oCfg.TraceStatements = statements

	p, err := factory.CreateTraces(b.Context(), processortest.NewNopSettings(metadata.Type), oCfg, consumertest.NewNop())
	require.NoError(b, err)

	input, err := golden.ReadTraces(filepath.Join("testdata", "traces", "input.yaml"))
	require.NoError(b, err)

	for b.Loop() {
		cp := ptrace.NewTraces()
		input.CopyTo(cp)
		require.NoError(b, p.ConsumeTraces(b.Context(), cp))
	}
}

func BenchmarkTraces_Baseline_Copy(b *testing.B) {
	input, err := golden.ReadTraces(filepath.Join("testdata", "traces", "input.yaml"))
	require.NoError(b, err)

	for b.Loop() {
		cp := ptrace.NewTraces()
		input.CopyTo(cp)
	}
}

func BenchmarkTraces_Merge_WithoutWildcards(b *testing.B) {
	benchmarkTraces(b, []config.ContextStatements{
		{
			Mode:    config.Merge,
			Context: "span",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Merge,
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
	})
}

func BenchmarkTraces_Merge_WithWildcards(b *testing.B) {
	benchmarkTraces(b, []config.ContextStatements{
		{
			Mode:    config.Merge,
			Context: "span",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Merge,
			Context: "resource",
			Attributes: []string{
				"k8s.*",
				"service.*",
				"deployment.environment.name",
				"host.name",
			},
		},
	})
}

func BenchmarkTraces_Move_WithoutWildcards(b *testing.B) {
	benchmarkTraces(b, []config.ContextStatements{
		{
			Mode:    config.Move,
			Context: "span",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Move,
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
	})
}

func BenchmarkTraces_Move_WithWildcards(b *testing.B) {
	benchmarkTraces(b, []config.ContextStatements{
		{
			Mode:    config.Move,
			Context: "span",
			Attributes: []string{
				"team",
			},
		},
		{
			Mode:    config.Move,
			Context: "resource",
			Attributes: []string{
				"k8s.*",
				"service.*",
				"deployment.environment.name",
				"host.name",
			},
		},
	})
}
