package constant

import (
	"testing"

	"ttpos-server-go/app/constant/jwt"
)

func TestMapJwtSourceToSaleBillSource(t *testing.T) {
	tests := []struct {
		name      string
		jwtSource string
		want      uint
	}{
		{
			name:      "cashier",
			jwtSource: jwt.SourceCashier,
			want:      SaleBillSourceCashier,
		},
		{
			name:      "assistant",
			jwtSource: jwt.SourceAssistant,
			want:      SaleBillSourceAssistant,
		},
		{
			name:      "tablet",
			jwtSource: jwt.SourceTablet,
			want:      SaleBillSourceTablet,
		},
		{
			name:      "h5",
			jwtSource: jwt.SourceH5,
			want:      SaleBillSourceH5,
		},
		{
			name:      "unknown",
			jwtSource: "unknown",
			want:      SaleBillSourceDefault,
		},
		{
			name:      "empty",
			jwtSource: "",
			want:      SaleBillSourceDefault,
		},
		{
			name:      "kitchen",
			jwtSource: jwt.SourceKitchen,
			want:      SaleBillSourceDefault,
		},
		{
			name:      "member",
			jwtSource: jwt.SourceMember,
			want:      SaleBillSourceMember,
		},
		{
			name:      "kiosk",
			jwtSource: jwt.SourceKiosk,
			want:      SaleBillSourceKiosk,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapJwtSourceToSaleBillSource(tt.jwtSource); got != tt.want {
				t.Errorf("MapJwtSourceToSaleBillSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeClientVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "normal",
			version: "2.10.0",
			want:    "2.10.0",
		},
		{
			name:    "with_v_lowercase",
			version: "v2.10.0",
			want:    "2.10.0",
		},
		{
			name:    "with_V_uppercase",
			version: "V2.10.0",
			want:    "2.10.0",
		},
		{
			name:    "with_space",
			version: " 2.10.0 ",
			want:    "2.10.0",
		},
		{
			name:    "empty",
			version: "",
			want:    "",
		},
		{
			name:    "only_v",
			version: "v",
			want:    "",
		},
		{
			name:    "v_with_space",
			version: " v2.10.0 ",
			want:    "2.10.0",
		},
		{
			name:    "V_with_space",
			version: " V2.10.0 ",
			want:    "2.10.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeClientVersion(tt.version); got != tt.want {
				t.Errorf("NormalizeClientVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
