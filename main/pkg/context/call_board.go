package context

import "context"

type ctxKeyCompanyUuid struct{}

type ctxKeyDeviceId struct{}

func GetDeviceId(c context.Context) string {
	return c.Value(ctxKeyDeviceId{}).(string)
}

func GetCompanyUuid(c context.Context) uint64 {
	return c.Value(ctxKeyCompanyUuid{}).(uint64)
}

func WithDeviceId(c context.Context, deviceId string) context.Context {
	return context.WithValue(c, ctxKeyDeviceId{}, deviceId)
}

func WithCompanyUuidV2(c context.Context, companyUuid uint64) context.Context {
	return context.WithValue(c, ctxKeyCompanyUuid{}, companyUuid)
}
