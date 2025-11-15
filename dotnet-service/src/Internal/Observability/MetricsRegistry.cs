using System;
using DotnetService.Internal.Application.Services;
using Prometheus;

namespace DotnetService.Internal.Observability;

public static class MetricsRegistry
{
    private static readonly object SyncRoot = new();
    private static bool _initialized;
    private static bool _enabled;
    private static string _serviceName = "dotnet-service";

    private static Counter? _httpRequestsTotal;
    private static Histogram? _httpRequestDurationSeconds;
    private static Counter? _cacheHitsTotal;
    private static Counter? _cacheMissesTotal;
    private static Gauge? _writeBehindLagSeconds;
    private static Gauge? _writeBehindQueueLength;
    private static Gauge? _writeBehindBatchSize;
    private static Histogram? _writeBehindBatchDurationSeconds;
    private static Counter? _writeBehindErrorsTotal;

    public static void Initialize(string serviceName, bool enabled)
    {
        if (_initialized)
        {
            _enabled = enabled;
            return;
        }
        lock (SyncRoot)
        {
            if (_initialized)
            {
                _enabled = enabled;
                return;
            }
            _serviceName = serviceName;
            _enabled = enabled;
            _httpRequestsTotal = Metrics.CreateCounter("http_requests_total", "Total HTTP requests handled by the API", new CounterConfiguration
            {
                LabelNames = new[] { "service", "method", "route", "status" }
            });
            _httpRequestDurationSeconds = Metrics.CreateHistogram("http_request_duration_seconds", "Duration of HTTP handlers in seconds", new HistogramConfiguration
            {
                LabelNames = new[] { "service", "method", "route" },
                Buckets = Histogram.ExponentialBuckets(0.005, 2, 10)
            });
            _cacheHitsTotal = Metrics.CreateCounter("cache_hits_total", "Total number of cache hits", new CounterConfiguration
            {
                LabelNames = new[] { "service", "status" }
            });
            _cacheMissesTotal = Metrics.CreateCounter("cache_misses_total", "Total number of cache misses or bypasses", new CounterConfiguration
            {
                LabelNames = new[] { "service", "status" }
            });
            _writeBehindLagSeconds = Metrics.CreateGauge("write_behind_lag_seconds", "Age in seconds of the oldest message in the latest batch", new GaugeConfiguration
            {
                LabelNames = new[] { "service" }
            });
            _writeBehindQueueLength = Metrics.CreateGauge("write_behind_queue_length", "Current Redis stream length for write-behind operations", new GaugeConfiguration
            {
                LabelNames = new[] { "service" }
            });
            _writeBehindBatchSize = Metrics.CreateGauge("write_behind_batch_size", "Number of events processed in the latest batch", new GaugeConfiguration
            {
                LabelNames = new[] { "service" }
            });
            _writeBehindBatchDurationSeconds = Metrics.CreateHistogram("write_behind_batch_duration_seconds", "Processing latency for write-behind batches", new HistogramConfiguration
            {
                LabelNames = new[] { "service" },
                Buckets = Histogram.ExponentialBuckets(0.01, 2, 10)
            });
            _writeBehindErrorsTotal = Metrics.CreateCounter("write_behind_errors_total", "Total write-behind errors encountered", new CounterConfiguration
            {
                LabelNames = new[] { "service" }
            });
            _initialized = true;
        }
    }

    private static bool Enabled => _initialized && _enabled;

    public static void ObserveHttpRequest(string method, string route, int statusCode, TimeSpan duration)
    {
        if (!Enabled)
        {
            return;
        }
        _httpRequestsTotal!
            .WithLabels(_serviceName, method, route, statusCode.ToString())
            .Inc();
        _httpRequestDurationSeconds!
            .WithLabels(_serviceName, method, route)
            .Observe(duration.TotalSeconds);
    }

    public static void RecordCacheStatus(CacheStatus status)
    {
        if (!Enabled)
        {
            return;
        }
        var normalized = status.ToString().ToLowerInvariant();
        if (status is CacheStatus.Fresh or CacheStatus.Stale)
        {
            _cacheHitsTotal!.WithLabels(_serviceName, normalized).Inc();
        }
        else
        {
            _cacheMissesTotal!.WithLabels(_serviceName, normalized).Inc();
        }
    }

    public static void RecordWriteBehindBatch(int batchSize, TimeSpan duration, double lagSeconds, long queueLength)
    {
        if (!Enabled || batchSize <= 0)
        {
            return;
        }
        _writeBehindBatchSize!.WithLabels(_serviceName).Set(batchSize);
        _writeBehindBatchDurationSeconds!.WithLabels(_serviceName).Observe(duration.TotalSeconds);
        if (lagSeconds >= 0)
        {
            _writeBehindLagSeconds!.WithLabels(_serviceName).Set(lagSeconds);
        }
        if (queueLength >= 0)
        {
            _writeBehindQueueLength!.WithLabels(_serviceName).Set(queueLength);
        }
    }

    public static void RecordWriteBehindError()
    {
        if (!Enabled)
        {
            return;
        }
        _writeBehindErrorsTotal!.WithLabels(_serviceName).Inc();
    }
}
