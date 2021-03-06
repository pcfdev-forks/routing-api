// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"code.cloudfoundry.org/routing-api/metrics"
)

type FakePartialStatsdClient struct {
	GaugeDeltaStub        func(stat string, value int64, rate float32) error
	gaugeDeltaMutex       sync.RWMutex
	gaugeDeltaArgsForCall []struct {
		stat  string
		value int64
		rate  float32
	}
	gaugeDeltaReturns struct {
		result1 error
	}
	GaugeStub        func(stat string, value int64, rate float32) error
	gaugeMutex       sync.RWMutex
	gaugeArgsForCall []struct {
		stat  string
		value int64
		rate  float32
	}
	gaugeReturns struct {
		result1 error
	}
}

func (fake *FakePartialStatsdClient) GaugeDelta(stat string, value int64, rate float32) error {
	fake.gaugeDeltaMutex.Lock()
	fake.gaugeDeltaArgsForCall = append(fake.gaugeDeltaArgsForCall, struct {
		stat  string
		value int64
		rate  float32
	}{stat, value, rate})
	fake.gaugeDeltaMutex.Unlock()
	if fake.GaugeDeltaStub != nil {
		return fake.GaugeDeltaStub(stat, value, rate)
	} else {
		return fake.gaugeDeltaReturns.result1
	}
}

func (fake *FakePartialStatsdClient) GaugeDeltaCallCount() int {
	fake.gaugeDeltaMutex.RLock()
	defer fake.gaugeDeltaMutex.RUnlock()
	return len(fake.gaugeDeltaArgsForCall)
}

func (fake *FakePartialStatsdClient) GaugeDeltaArgsForCall(i int) (string, int64, float32) {
	fake.gaugeDeltaMutex.RLock()
	defer fake.gaugeDeltaMutex.RUnlock()
	return fake.gaugeDeltaArgsForCall[i].stat, fake.gaugeDeltaArgsForCall[i].value, fake.gaugeDeltaArgsForCall[i].rate
}

func (fake *FakePartialStatsdClient) GaugeDeltaReturns(result1 error) {
	fake.GaugeDeltaStub = nil
	fake.gaugeDeltaReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakePartialStatsdClient) Gauge(stat string, value int64, rate float32) error {
	fake.gaugeMutex.Lock()
	fake.gaugeArgsForCall = append(fake.gaugeArgsForCall, struct {
		stat  string
		value int64
		rate  float32
	}{stat, value, rate})
	fake.gaugeMutex.Unlock()
	if fake.GaugeStub != nil {
		return fake.GaugeStub(stat, value, rate)
	} else {
		return fake.gaugeReturns.result1
	}
}

func (fake *FakePartialStatsdClient) GaugeCallCount() int {
	fake.gaugeMutex.RLock()
	defer fake.gaugeMutex.RUnlock()
	return len(fake.gaugeArgsForCall)
}

func (fake *FakePartialStatsdClient) GaugeArgsForCall(i int) (string, int64, float32) {
	fake.gaugeMutex.RLock()
	defer fake.gaugeMutex.RUnlock()
	return fake.gaugeArgsForCall[i].stat, fake.gaugeArgsForCall[i].value, fake.gaugeArgsForCall[i].rate
}

func (fake *FakePartialStatsdClient) GaugeReturns(result1 error) {
	fake.GaugeStub = nil
	fake.gaugeReturns = struct {
		result1 error
	}{result1}
}

var _ metrics.PartialStatsdClient = new(FakePartialStatsdClient)
