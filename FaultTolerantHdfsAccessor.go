// Copyright (c) Microsoft. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.
package main

import (
	"os"
)

// Adds automatic retry capability to HdfsAccessor with respect to RetryPolicy
type FaultTolerantHdfsAccessor struct {
	Impl        HdfsAccessor
	RetryPolicy *RetryPolicy
}

var _ HdfsAccessor = (*FaultTolerantHdfsAccessor)(nil) // ensure FaultTolerantHdfsAccessor implements HdfsAccessor

// Creates an instance of FaultTolerantHdfsAccessor
func NewFaultTolerantHdfsAccessor(impl HdfsAccessor, retryPolicy *RetryPolicy) *FaultTolerantHdfsAccessor {
	return &FaultTolerantHdfsAccessor{
		Impl:        impl,
		RetryPolicy: retryPolicy}
}

// Ensures HDFS accessor is connected to the HDFS name node
func (this *FaultTolerantHdfsAccessor) EnsureConnected() error {
	op := this.RetryPolicy.StartOperation()
	for {
		err := this.Impl.EnsureConnected()
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("Connect: %s", err) {
			return err
		}
	}
}

// Opens HDFS file for reading
func (this *FaultTolerantHdfsAccessor) OpenRead(path string) (HdfsReader, error) {
	op := this.RetryPolicy.StartOperation()
	for {
		result, err := this.Impl.OpenRead(path)
		if err == nil {
			// wrapping returned HdfsReader with FaultTolerantHdfsReader
			return NewFaultTolerantHdfsReader(path, result, this.Impl, this.RetryPolicy), nil
		}
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("[%s] OpenRead: %s", path, err) {
			return nil, err
		}
	}
}

// Opens HDFS file for efficient concurrent random access. Returns file size and RandomAccessHdfsReader interface
func (this *FaultTolerantHdfsAccessor) OpenReadForRandomAccess(path string) (RandomAccessHdfsReader, uint64, error) {
	op := this.RetryPolicy.StartOperation()
	for {
		reader, size, err := this.Impl.OpenReadForRandomAccess(path)
		if err == nil {
			return reader, size, nil
		}
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("[%s] OpenReadForRandomAccess: %s", path, err) {
			return nil, 0, err
		}
	}
}

// Opens HDFS file for writing
func (this *FaultTolerantHdfsAccessor) OpenWrite(path string) (HdfsWriter, error) {
	op := this.RetryPolicy.StartOperation()
	for {
		result, err := this.Impl.OpenWrite(path)
		if err == nil {
			// wrapping returned HdfsWriter with FaultTolerantHdfsReader
			return &FaultTolerantHdfsWriter{Impl: result}, nil
		}
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("[%s] OpenWrite: %s", path, err) {
			return nil, err
		}
	}
}

// Enumerates HDFS directory
func (this *FaultTolerantHdfsAccessor) ReadDir(path string) ([]Attrs, error) {
	op := this.RetryPolicy.StartOperation()
	for {
		result, err := this.Impl.ReadDir(path)
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("[%s] ReadDir: %s", path, err) {
			return result, err
		}
	}
}

// Retrieves file/directory attributes
func (this *FaultTolerantHdfsAccessor) Stat(path string) (Attrs, error) {
	op := this.RetryPolicy.StartOperation()
	for {
		result, err := this.Impl.Stat(path)
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("[%s] Stat: %s", path, err) {
			return result, err
		}
	}
}

// Creates a directory
func (this *FaultTolerantHdfsAccessor) Mkdir(path string, mode os.FileMode) error {
	op := this.RetryPolicy.StartOperation()
	for {
		err := this.Impl.Mkdir(path, mode)
		if IsSuccessOrBenignError(err) || !op.ShouldRetry("[%s] Mkdir %s: %s", path, mode, err) {
			return err
		}
	}
}