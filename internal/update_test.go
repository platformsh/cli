package internal

import (
	"reflect"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	type args struct {
		a *Version
		b *Version
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "major check", args: args{
			a: &Version{VersionParts: [3]int{2, 1, 1}},
			b: &Version{VersionParts: [3]int{1, 1, 1}},
		}, want: 1},
		{name: "major check reverse", args: args{
			a: &Version{VersionParts: [3]int{1, 1, 1}},
			b: &Version{VersionParts: [3]int{2, 1, 1}},
		}, want: -1},

		{name: "minor check", args: args{
			a: &Version{VersionParts: [3]int{1, 2, 1}},
			b: &Version{VersionParts: [3]int{1, 1, 1}},
		}, want: 1},
		{name: "minor check reverse", args: args{
			a: &Version{VersionParts: [3]int{1, 1, 1}},
			b: &Version{VersionParts: [3]int{1, 2, 1}},
		}, want: -1},

		{name: "patch check", args: args{
			a: &Version{VersionParts: [3]int{1, 1, 2}},
			b: &Version{VersionParts: [3]int{1, 1, 1}},
		}, want: 1},
		{name: "patch check reverse", args: args{
			a: &Version{VersionParts: [3]int{1, 1, 1}},
			b: &Version{VersionParts: [3]int{1, 1, 2}},
		}, want: -1},

		{name: "beta alpha check", args: args{
			a: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"beta"}},
			b: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"alpha"}},
		}, want: 1},
		{name: "beta alpha check reverse", args: args{
			a: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"alpha"}},
			b: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"beta"}},
		}, want: -1},

		{name: "beta check", args: args{
			a: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"beta", "2"}},
			b: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"beta", "1"}},
		}, want: 1},
		{name: "beta check reverse", args: args{
			a: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"beta", "1"}},
			b: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"beta", "2"}},
		}, want: -1},

		{name: "numeric check", args: args{
			a: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"beta", "02"}},
			b: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"beta", "1"}},
		}, want: 1},
		{name: "numeric check reverse", args: args{
			a: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"beta", "1"}},
			b: &Version{VersionParts: [3]int{1, 1, 1}, PreReleaseParts: []string{"beta", "02"}},
		}, want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareVersions(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("CompareVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		version string
		want    *Version
		wantErr bool
	}{
		{
			version: "0.1.2",
			want:    &Version{VersionParts: [3]int{0, 1, 2}},
			wantErr: false,
		},
		{
			version: "0.01.02",
			want:    &Version{VersionParts: [3]int{0, 1, 2}},
			wantErr: false,
		},
		{
			version: "0.01.02-beta.1",
			want: &Version{VersionParts: [3]int{0, 1, 2},
				PreReleaseParts: []string{"beta", "1"}},
			wantErr: false,
		},
		{
			version: "00.01.02-beta.001.pre",
			want: &Version{VersionParts: [3]int{0, 1, 2},
				PreReleaseParts: []string{"beta", "001", "pre"}},
			wantErr: false,
		},

		{version: "01.02-beta.001.pre", want: nil, wantErr: true},
		{version: "222.01", want: nil, wantErr: true},
		{version: "", want: nil, wantErr: true},
		{version: "ab.2.3", want: nil, wantErr: true},
		{version: "2ab.2.3", want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got, err := ParseVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
