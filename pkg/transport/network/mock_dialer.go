// Code generated by mockery v1.0.0. DO NOT EDIT.

package network

import (
	context "context"
	net "net"

	cipher "github.com/skycoin/dmsg/cipher"
	mock "github.com/stretchr/testify/mock"
)

// MockDialer is an autogenerated mock type for the Dialer type
type MockDialer struct {
	mock.Mock
}

// Dial provides a mock function with given fields: ctx, remote, port
func (_m *MockDialer) Dial(ctx context.Context, remote cipher.PubKey, port uint16) (net.Conn, error) {
	ret := _m.Called(ctx, remote, port)

	var r0 net.Conn
	if rf, ok := ret.Get(0).(func(context.Context, cipher.PubKey, uint16) net.Conn); ok {
		r0 = rf(ctx, remote, port)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(net.Conn)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, cipher.PubKey, uint16) error); ok {
		r1 = rf(ctx, remote, port)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Type provides a mock function with given fields:
func (_m *MockDialer) Type() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
