package urls

import "testing"

func TestGetDomainSuffix(t *testing.T) {
	wantGoogle := "google.com"
	tests := []struct {
		name    string
		arg     string
		want    string
		wantErr bool
	}{
		{name: "root domain", arg: "https://google.com", want: wantGoogle},
		{name: "sub domain", arg: "https://www.google.com", want: wantGoogle},
		{name: "multi sub domain", arg: "https://www2.www.google.com", want: wantGoogle},
		{name: "nonstandard port", arg: "https://www2.www.google.com:8080", want: wantGoogle},
		{name: "path", arg: "https://www2.www.google.com:8080/example/2", want: wantGoogle},
		{name: "no scheme", arg: "www.google.com", want: "www.google.com"},
		{name: "mistake scheme", arg: "httpss://www.google.com", want: "google.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDomainSuffix(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDomainSuffix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDomainSuffix() got = %v, want %v", got, tt.want)
			}
		})
	}
}
