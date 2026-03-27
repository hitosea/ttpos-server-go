package service

import (
	"testing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"

	pkgerrors "github.com/pkg/errors"
)

func TestGetHeadquarterUuid(t *testing.T) {
	t.Parallel()

	srv := &companyAreaSrv{}

	tests := []struct {
		name            string
		companyUuid     uint64
		companySetting  model.CompanySetting
		expectedHqUuid  uint64
	}{
		{
			name:        "总部调用_返回自身UUID",
			companyUuid: 1001,
			companySetting: model.CompanySetting{
				HeadquarterUuid:        9999, // 不应使用此值
				ErpnextSiteCode:        "site1",
				ErpnextCompanyAbbr:     "HQ",
				ErpnextHeadquarterAbbr: "HQ", // CompanyAbbr == HeadquarterAbbr → IsHeadquarter
			},
			expectedHqUuid: 1001,
		},
		{
			name:        "子店调用_返回HeadquarterUuid",
			companyUuid: 2001,
			companySetting: model.CompanySetting{
				HeadquarterUuid:        1001,
				ErpnextSiteCode:        "site1",
				ErpnextCompanyAbbr:     "SUB1",
				ErpnextHeadquarterAbbr: "HQ", // CompanyAbbr != HeadquarterAbbr → IsSubShop
			},
			expectedHqUuid: 1001,
		},
		{
			name:        "散户_SiteCode为空_返回HeadquarterUuid",
			companyUuid: 3001,
			companySetting: model.CompanySetting{
				HeadquarterUuid:        0,
				ErpnextSiteCode:        "", // IsTtposSite → true
				ErpnextCompanyAbbr:     "",
				ErpnextHeadquarterAbbr: "",
			},
			expectedHqUuid: 0,
		},
		{
			name:        "散户_SiteCode为1_返回HeadquarterUuid",
			companyUuid: 4001,
			companySetting: model.CompanySetting{
				HeadquarterUuid:        0,
				ErpnextSiteCode:        "1", // IsTtposSite → true
				ErpnextCompanyAbbr:     "A",
				ErpnextHeadquarterAbbr: "A",
			},
			expectedHqUuid: 0, // 散户 IsTtposSite=true → IsHeadquarter=false → 走 HeadquarterUuid 分支
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.NewContext(context.WithCompanyUuid(tt.companyUuid))
			ctx.SetCompanySetting(tt.companySetting)

			got := srv.getHeadquarterUuid(ctx)
			if got != tt.expectedHqUuid {
				t.Errorf("getHeadquarterUuid() = %d, want %d", got, tt.expectedHqUuid)
			}
		})
	}
}

func TestRequireHeadquarter(t *testing.T) {
	t.Parallel()

	srv := &companyAreaSrv{}

	tests := []struct {
		name           string
		companySetting model.CompanySetting
		wantErr        bool
		wantCode       int
	}{
		{
			name: "总部_通过",
			companySetting: model.CompanySetting{
				ErpnextSiteCode:        "site1",
				ErpnextCompanyAbbr:     "HQ",
				ErpnextHeadquarterAbbr: "HQ",
			},
			wantErr: false,
		},
		{
			name: "子店_拒绝",
			companySetting: model.CompanySetting{
				ErpnextSiteCode:        "site1",
				ErpnextCompanyAbbr:     "SUB1",
				ErpnextHeadquarterAbbr: "HQ",
			},
			wantErr:  true,
			wantCode: constant.CodeAccessDenied,
		},
		{
			name: "散户_拒绝",
			companySetting: model.CompanySetting{
				ErpnextSiteCode:        "",
				ErpnextCompanyAbbr:     "",
				ErpnextHeadquarterAbbr: "",
			},
			wantErr:  true,
			wantCode: constant.CodeAccessDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.NewContext()
			ctx.SetCompanySetting(tt.companySetting)

			err := srv.requireHeadquarter(ctx)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var appErr errors.AppError
				if !pkgerrors.As(err, &appErr) {
					t.Fatalf("expected AppError, got %T", err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("error code = %d, want %d", appErr.Code, tt.wantCode)
				}
			} else {
				if err != nil {
					t.Errorf("expected nil, got %v", err)
				}
			}
		})
	}
}
