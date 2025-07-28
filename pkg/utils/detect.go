package utils

import (
	"context"
	"net"
	"strings"
)

func DetectRecordType(input string) string {
	input = strings.TrimSpace(input)
	resolver := net.Resolver{}
	ctx := context.Background()

	if ip := net.ParseIP(input); ip != nil {
		return "ip"
	}
	if _, err := resolver.LookupNS(ctx, input); err == nil {
		return "ns"
	}
	if _, err := resolver.LookupCNAME(ctx, input); err == nil {
		return "cname"
	}
	if _, err := resolver.LookupTXT(ctx, input); err == nil {
		return "txt"
	}
	if _, err := resolver.LookupMX(ctx, input); err == nil {
		return "mx"
	}
	return ""
}
