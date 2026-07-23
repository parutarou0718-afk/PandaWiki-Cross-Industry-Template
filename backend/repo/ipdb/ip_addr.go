package ipdb

import (
	"context"
	"net"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/store/ipdb"
	"github.com/chaitin/panda-wiki/utils"
)

type IPAddressRepo struct {
	ipdb   *ipdb.IPDB
	logger *log.Logger
}

func NewIPAddressRepo(ipdb *ipdb.IPDB, logger *log.Logger) *IPAddressRepo {
	return &IPAddressRepo{ipdb: ipdb, logger: logger.WithModule("repo.ipdb.ip_addr")}
}

func (r *IPAddressRepo) GetIPAddress(ctx context.Context, ip string) (*domain.IPAddress, error) {
	if ip == "" || net.ParseIP(ip) == nil {
		return &domain.IPAddress{
			IP:       ip,
			Country:  "无效地址",
			Province: "无效地址",
			City:     "无效地址",
		}, nil
	}
	if utils.IsPrivateOrReservedIP(ip) {
		return &domain.IPAddress{
			IP:       ip,
			Country:  "保留地址",
			Province: "保留地址",
			City:     "保留地址",
		}, nil
	}
	info, err := r.ipdb.Lookup(ip)
	if err != nil {
		r.logger.Error("failed to lookup ip address", log.Any("error", err), log.String("ip", ip))
		return &domain.IPAddress{
			IP:       ip,
			Country:  "未知地址",
			Province: "未知地址",
			City:     "未知地址",
		}, nil
	}
	return info, nil
}

func (r *IPAddressRepo) GetIPAddresses(ctx context.Context, ips []string) (map[string]*domain.IPAddress, error) {
	ipAddresses := make(map[string]*domain.IPAddress, len(ips))
	for _, ip := range ips {
		info, err := r.GetIPAddress(ctx, ip)
		if err != nil {
			return nil, err
		}
		ipAddresses[ip] = info
	}
	return ipAddresses, nil
}
