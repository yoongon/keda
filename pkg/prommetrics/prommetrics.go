/*
Copyright 2022 The KEDA Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package prommetrics

import (
	"runtime"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/kedacore/keda/v2/version"
)

var log = logf.Log.WithName("prometheus_server")

const (
	ClusterTriggerAuthenticationResource = "cluster_trigger_authentication"
	TriggerAuthenticationResource        = "trigger_authentication"
	ScaledObjectResource                 = "scaled_object"
	ScaledJobResource                    = "scaled_job"

	DefaultPromMetricsNamespace = "keda"
)

var (
	scaledObjectMetricLabels = []string{"namespace", "metric", "scaledObject", "scaler", "scalerIndex"}
	scaledJobmetricLabels    = []string{"namespace", "metric", "scaledJob", "scaler", "scalerIndex"}
	buildInfo                = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: DefaultPromMetricsNamespace,
			Name:      "build_info",
			Help:      "A metric with a constant '1' value labeled by version, git_commit and goversion from which KEDA was built.",
		},
		[]string{"version", "git_commit", "goversion", "goos", "goarch"},
	)
	scalerErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "scaler",
			Name:      "errors_total",
			Help:      "Total number of errors for all scalers",
		},
		[]string{},
	)
	scaledObjectScalerMetricsValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "scaler",
			Name:      "scaledobject_metrics_value",
			Help:      "Metric Value used for HPA",
		},
		scaledObjectMetricLabels,
	)
	scaledJobScalerMetricsValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "scaler",
			Name:      "scaledjob_metrics_value",
			Help:      "Metric Value used for HPA",
		},
		scaledJobmetricLabels,
	)
	scaledObjectScalerMetricsLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "scaler",
			Name:      "scaledobject_metrics_latency",
			Help:      "Scaler Metrics Latency",
		},
		scaledObjectMetricLabels,
	)
	scaledJobScalerMetricsLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "scaler",
			Name:      "scaledjob_metrics_latency",
			Help:      "Scaler Metrics Latency",
		},
		scaledJobmetricLabels,
	)
	scaledObjectScalerActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "scaler",
			Name:      "scaledobject_active",
			Help:      "Activity of a Scaler Metric",
		},
		scaledObjectMetricLabels,
	)
	scaledJobScalerActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "scaler",
			Name:      "scaledjob_active",
			Help:      "Activity of a Scaler Metric",
		},
		scaledJobmetricLabels,
	)
	scaledObjectScalerErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "scaler",
			Name:      "scaledobject_errors",
			Help:      "Number of scaler errors",
		},
		scaledObjectMetricLabels,
	)
	scaledJobScalerErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "scaler",
			Name:      "scaledjob_errors",
			Help:      "Number of scaler errors",
		},
		scaledJobmetricLabels,
	)
	scaledObjectErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "scaled_object",
			Name:      "errors",
			Help:      "Number of scaled object errors",
		},
		[]string{"namespace", "scaledObject"},
	)
	scaledJobErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "scaled_job",
			Name:      "errors",
			Help:      "Number of scaled object errors",
		},
		[]string{"namespace", "scaledJob"},
	)

	triggerTotalsGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "trigger",
			Name:      "totals",
		},
		[]string{"type"},
	)

	crdTotalsGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "resource",
			Name:      "totals",
		},
		[]string{"type", "namespace"},
	)

	internalLoopLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: DefaultPromMetricsNamespace,
			Subsystem: "internal_scale_loop",
			Name:      "latency",
			Help:      "Internal latency of ScaledObject/ScaledJob loop execution",
		},
		[]string{"namespace", "type", "resource"},
	)
)

func init() {
	metrics.Registry.MustRegister(scalerErrorsTotal)
	metrics.Registry.MustRegister(internalLoopLatency)

	metrics.Registry.MustRegister(scaledObjectScalerMetricsValue)
	metrics.Registry.MustRegister(scaledObjectScalerMetricsLatency)
	metrics.Registry.MustRegister(scaledObjectScalerActive)
	metrics.Registry.MustRegister(scaledObjectScalerErrors)
	metrics.Registry.MustRegister(scaledObjectErrors)

	metrics.Registry.MustRegister(scaledJobScalerMetricsValue)
	metrics.Registry.MustRegister(scaledJobScalerMetricsLatency)
	metrics.Registry.MustRegister(scaledJobScalerActive)
	metrics.Registry.MustRegister(scaledJobScalerErrors)
	metrics.Registry.MustRegister(scaledJobErrors)

	metrics.Registry.MustRegister(triggerTotalsGaugeVec)
	metrics.Registry.MustRegister(crdTotalsGaugeVec)
	metrics.Registry.MustRegister(buildInfo)

	RecordBuildInfo()
}

// RecordScalerMetric create a measurement of the external metric used by the HPA
func RecordScalerMetric(namespace string, scaledResource string, scaler string, scalerIndex int, metric string, value float64, resourceType string) {
	labels := getLabels(namespace, scaledResource, scaler, scalerIndex, metric, resourceType)
	switch resourceType {
	case ScaledObjectResource:
		scaledObjectScalerMetricsValue.With(labels).Set(value)
	case ScaledJobResource:
		scaledJobScalerMetricsValue.With(labels).Set(value)
	}
}

// RecordScalerLatency create a measurement of the latency to external metric
func RecordScalerLatency(namespace string, scaledResource string, scaler string, scalerIndex int, metric string, value float64, resourceType string) {
	labels := getLabels(namespace, scaledResource, scaler, scalerIndex, metric, resourceType)
	switch resourceType {
	case ScaledObjectResource:
		scaledObjectScalerMetricsLatency.With(labels).Set(value)
	case ScaledJobResource:
		scaledJobScalerMetricsLatency.With(labels).Set(value)
	}
}

// RecordScaledObjectLatency create a measurement of the latency executing scalable object loop
func RecordScalableObjectLatency(namespace string, name string, isScaledObject bool, value float64) {
	resourceType := "scaledjob"
	if isScaledObject {
		resourceType = "scaledobject"
	}
	internalLoopLatency.WithLabelValues(namespace, resourceType, name).Set(value)
}

// RecordScalerActive create a measurement of the activity of the scaler
func RecordScalerActive(namespace string, scaledResource string, scaler string, scalerIndex int, metric string, active bool, resourceType string) {
	activeVal := 0
	if active {
		activeVal = 1
	}
	labels := getLabels(namespace, scaledResource, scaler, scalerIndex, metric, resourceType)
	switch resourceType {
	case ScaledObjectResource:
		scaledObjectScalerActive.With(labels).Set(float64(activeVal))
	case ScaledJobResource:
		scaledJobScalerActive.With(labels).Set(float64(activeVal))
	}
}

// RecordScalerError counts the number of errors occurred in trying to get an external metric used by the HPA
func RecordScalerError(namespace string, scaledResource string, scaler string, scalerIndex int, metric string, err error, resourceType string) {
	labels := getLabels(namespace, scaledResource, scaler, scalerIndex, metric, resourceType)
	switch resourceType {
	case ScaledObjectResource:
		if err != nil {
			scaledObjectScalerErrors.With(labels).Inc()
			RecordScaledObjectError(namespace, scaledResource, err, resourceType)
			scalerErrorsTotal.With(prometheus.Labels{}).Inc()
			break
		}
		// initialize metric with 0 if not already set
		_, errScaler := scaledObjectScalerErrors.GetMetricWith(labels)
		if errScaler != nil {
			log.Error(errScaler, "Unable to write to metrics to Prometheus Server: %v")
		}
	case ScaledJobResource:
		if err != nil {
			scaledJobScalerErrors.With(labels).Inc()
			RecordScaledObjectError(namespace, scaledResource, err, resourceType)
			scalerErrorsTotal.With(prometheus.Labels{}).Inc()
			break
		}
		// initialize metric with 0 if not already set
		_, errScaler := scaledJobScalerErrors.GetMetricWith(labels)
		if errScaler != nil {
			log.Error(errScaler, "Unable to write to metrics to Prometheus Server: %v")
		}
	}
}

// RecordScaleObjectError counts the number of errors with the scaled object
func RecordScaledObjectError(namespace string, scaledResource string, err error, resourceType string) {
	switch resourceType {
	case ScaledObjectResource:
		labels := prometheus.Labels{"namespace": namespace, "scaledObject": scaledResource}
		if err != nil {
			scaledObjectErrors.With(labels).Inc()
			return
		}
		// initialize metric with 0 if not already set
		_, errScaledObject := scaledObjectErrors.GetMetricWith(labels)
		if errScaledObject != nil {
			log.Error(errScaledObject, "Unable to write to metrics to Prometheus Server: %v")
			return
		}
	case ScaledJobResource:
		labels := prometheus.Labels{"namespace": namespace, "scaledJob": scaledResource}
		if err != nil {
			scaledJobErrors.With(labels).Inc()
			return
		}
		// initialize metric with 0 if not already set
		_, errScaledJob := scaledJobErrors.GetMetricWith(labels)
		if errScaledJob != nil {
			log.Error(errScaledJob, "Unable to write to metrics to Prometheus Server: %v")
			return
		}
	}
}

// RecordBuildInfo publishes information about KEDA version and runtime info through an info metric (gauge).
func RecordBuildInfo() {
	buildInfo.WithLabelValues(version.Version, version.GitCommit, runtime.Version(), runtime.GOOS, runtime.GOARCH).Set(1)
}

func getLabels(namespace string, scaledResource string, scaler string, scalerIndex int, metric string, resourceType string) prometheus.Labels {
	switch resourceType {
	case ScaledObjectResource:
		return prometheus.Labels{"namespace": namespace, "scaledObject": scaledResource, "scaler": scaler, "scalerIndex": strconv.Itoa(scalerIndex), "metric": metric}
	case ScaledJobResource:
		return prometheus.Labels{"namespace": namespace, "scaledJob": scaledResource, "scaler": scaler, "scalerIndex": strconv.Itoa(scalerIndex), "metric": metric}
	}
	// Only two types(ScaledObject and ScaledJob) are currently supported. It cannot be reached here
	return nil
}

func IncrementTriggerTotal(triggerType string) {
	if triggerType != "" {
		triggerTotalsGaugeVec.WithLabelValues(triggerType).Inc()
	}
}

func DecrementTriggerTotal(triggerType string) {
	if triggerType != "" {
		triggerTotalsGaugeVec.WithLabelValues(triggerType).Dec()
	}
}

func IncrementCRDTotal(crdType, namespace string) {
	if namespace == "" {
		namespace = "default"
	}

	crdTotalsGaugeVec.WithLabelValues(crdType, namespace).Inc()
}

func DecrementCRDTotal(crdType, namespace string) {
	if namespace == "" {
		namespace = "default"
	}

	crdTotalsGaugeVec.WithLabelValues(crdType, namespace).Dec()
}
