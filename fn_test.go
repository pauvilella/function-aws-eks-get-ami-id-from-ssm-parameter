package main

import (
	"context"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"

	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
)

func TestRunFunction(t *testing.T) {
	type args struct {
		ctx context.Context
		req *fnv1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"AddTwoBuckets": {
			reason: "The Function should add two buckets to the desired composed resources",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Observed: &fnv1.State{
						Composite: &fnv1.Resource{
							// MustStructJSON is a handy way to provide mock
							// resources.
							Resource: resource.MustStructJSON(`{
								"apiVersion": "example.crossplane.io/v1alpha1",
								"kind": "XBuckets",
								"metadata": {
									"name": "test"
								},
								"spec": {
									"region": "us-east-2",
									"names": [
										"test-bucket-a",
										"test-bucket-b"
									]
								}
							}`),
						},
					},
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: &fnv1.ResponseMeta{Ttl: durationpb.New(60 * time.Second)},
					Desired: &fnv1.State{
						Resources: map[string]*fnv1.Resource{
							"xbuckets-test-bucket-a": {Resource: resource.MustStructJSON(`{
								"apiVersion": "s3.aws.upbound.io/v1beta1",
								"kind": "Bucket",
								"metadata": {
									"annotations": {
										"crossplane.io/external-name": "test-bucket-a"
									}
								},
								"spec": {
									"forProvider": {
										"region": "us-east-2"
									}
								},
								"status": {
									"observedGeneration": 0
								}
							}`)},
							"xbuckets-test-bucket-b": {Resource: resource.MustStructJSON(`{
								"apiVersion": "s3.aws.upbound.io/v1beta1",
								"kind": "Bucket",
								"metadata": {
									"annotations": {
										"crossplane.io/external-name": "test-bucket-b"
									}
								},
								"spec": {
									"forProvider": {
										"region": "us-east-2"
									}
								},
								"status": {
									"observedGeneration": 0
								}
							}`)},
						},
					},
					Conditions: []*fnv1.Condition{
						{
							Type:   "FunctionSuccess",
							Status: fnv1.Status_STATUS_CONDITION_TRUE,
							Reason: "Success",
							Target: fnv1.Target_TARGET_COMPOSITE.Enum(),
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)

			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}
}
