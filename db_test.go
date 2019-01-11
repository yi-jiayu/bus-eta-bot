package main

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

func NewStronglyConsistentContext() (context.Context, func(), error) {
	opts := aetest.Options{
		StronglyConsistentDatastore: true,
	}

	inst, err := aetest.NewInstance(&opts)
	if err != nil {
		return nil, nil, err
	}
	req, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		inst.Close()
		return nil, nil, err
	}
	ctx := appengine.NewContext(req)
	return ctx, func() {
		inst.Close()
	}, nil
}
