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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapJwtSourceToSaleBillSource(tt.jwtSource); got != tt.want {
				t.Errorf("MapJwtSourceToSaleBillSource() = %v, want %v", got, tt.want)
			}
		})
	}
}
