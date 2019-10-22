package client_test

import (
	"context"
	"errors"
	"fmt"
	"kube-trivy-exporter/pkg/client"
	"kube-trivy-exporter/pkg/domain"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTrivyClientDo(t *testing.T) {
	type in struct {
		first  context.Context
		second string
	}

	type want struct {
		first []domain.TrivyResponse
	}

	tests := []struct {
		receiver        *client.TrivyClient
		in              in
		want            want
		wantErrorString string
		optsFunction    func(interface{}) cmp.Option
	}{
		{
			&client.TrivyClient{
				Executor: func(context.Context, string, ...string) ([]byte, error) {
					return []byte(`[{"Target": "k8s.gcr.io/kube-addon-manager:v9.0.2 (debian 9.8)",
"Vulnerabilities":[{
"VulnerabilityID":"CVE-2011-3374",
"PkgName":"apt",
"InstalledVersion":"1.4.9",
"FixedVersion":"",
"Title":"",
"Description":"",
"Severity":"LOW",
"References":null
}]}]`), nil
				},
			},
			in{
				context.Background(),
				"dummy",
			},
			want{
				[]domain.TrivyResponse{
					{
						Target: "k8s.gcr.io/kube-addon-manager:v9.0.2 (debian 9.8)",
						Vulnerabilities: []domain.TrivyVulnerability{
							{
								VulnerabilityID:  "CVE-2011-3374",
								PkgName:          "apt",
								InstalledVersion: "1.4.9",
								FixedVersion:     "",
								Title:            "",
								Description:      "",
								Severity:         "LOW",
								References:       nil,
							},
						},
					},
				},
			},
			"",
			func(got interface{}) cmp.Option {
				return nil
			},
		},
		{
			&client.TrivyClient{
				Executor: func(context.Context, string, ...string) ([]byte, error) {
					return nil, errors.New("fake")
				},
			},
			in{
				context.Background(),
				"dummy",
			},
			want{
				nil,
			},
			"could not execute trivy command: fake",
			func(got interface{}) cmp.Option {
				return nil
			},
		},
		{
			&client.TrivyClient{
				Executor: func(context.Context, string, ...string) ([]byte, error) {
					return []byte("fake"), nil
				},
			},
			in{
				context.Background(),
				"dummy",
			},
			want{
				nil,
			},
			"could not parse trivy response: invalid character 'k' in literal false (expecting 'l')",
			func(got interface{}) cmp.Option {
				return nil
			},
		},
	}
	for _, tt := range tests {
		receiver := tt.receiver
		in := tt.in
		want := tt.want
		wantErrorString := tt.wantErrorString
		optsFunction := tt.optsFunction
		t.Run(fmt.Sprintf("%#v/%#v", receiver, in), func(t *testing.T) {
			t.Parallel()

			got, err := receiver.Do(in.first, in.second)
			if diff := cmp.Diff(want.first, got, optsFunction(got)); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
			if err != nil {
				gotErrorString := err.Error()
				if diff := cmp.Diff(wantErrorString, gotErrorString); diff != "" {
					t.Errorf("(-want +got):\n%s", diff)
				}
			}
		})
	}
}