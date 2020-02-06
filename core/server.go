package core

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"gitlab.127-0-0-1.fr/vx3r/wg-gen-web/model"
	"gitlab.127-0-0-1.fr/vx3r/wg-gen-web/storage"
	"gitlab.127-0-0-1.fr/vx3r/wg-gen-web/template"
	"gitlab.127-0-0-1.fr/vx3r/wg-gen-web/util"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReadServer object, create default one
func ReadServer() (*model.Server, error) {
	if !util.FileExists(filepath.Join(os.Getenv("WG_CONF_DIR"), "server.json")) {
		server := &model.Server{}

		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}
		server.PrivateKey = key.String()
		server.PublicKey = key.PublicKey().String()

		presharedKey, err := wgtypes.GenerateKey()
		if err != nil {
			return nil, err
		}
		server.PresharedKey = presharedKey.String()

		server.Name = "Created with default values"
		server.Endpoint = "wireguard.example.com:123"
		server.ListenPort = 51820
		server.Address = "fd9f:6666::10:6:6:1/112, 10.6.6.1/24"
		server.Dns = "fd9f::10:0:0:2, 10.0.0.2"
		server.PersistentKeepalive = 16
		server.PreUp = "echo WireGuard PreUp"
		server.PostUp = "echo WireGuard PostUp"
		server.PreDown = "echo WireGuard PreDown"
		server.PostDown = "echo WireGuard PostDown"
		server.Created = time.Now().UTC()
		server.Updated = server.Created

		err = storage.Serialize("server.json", server)
		if err != nil {
			return nil, err
		}
	}

	c, err := storage.Deserialize("server.json")
	if err != nil {
		return nil, err
	}

	return c.(*model.Server), nil
}

// UpdateServer keep private values from existing one
func UpdateServer(server *model.Server) (*model.Server, error) {
	current, err := storage.Deserialize("server.json")
	if err != nil {
		return nil, err
	}

	// check if server is valid
	errs := server.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("server validation error")
		}
		return nil, errors.New("failed to validate server")
	}

	server.PrivateKey = current.(*model.Server).PrivateKey
	server.PublicKey = current.(*model.Server).PublicKey
	server.PresharedKey = current.(*model.Server).PresharedKey
	server.Updated = time.Now().UTC()

	err = storage.Serialize("server.json", server)
	if err != nil {
		return nil, err
	}

	v, err := storage.Deserialize("server.json")
	if err != nil {
		return nil, err
	}
	server = v.(*model.Server)

	return server, UpdateServerConfigWg()
}

// UpdateServerConfigWg in wg format
func UpdateServerConfigWg() error {
	clients, err := ReadClients()
	if err != nil {
		return err
	}

	server, err := ReadServer()
	if err != nil {
		return err
	}

	_, err = template.DumpServerWg(clients, server)
	if err != nil {
		return err
	}

	return nil
}

// GetAllReservedIps the list of all reserved IPs, client and server
func GetAllReservedIps() ([]string, error) {
	clients, err := ReadClients()
	if err != nil {
		return nil, err
	}

	server, err := ReadServer()
	if err != nil {
		return nil, err
	}

	reserverIps := make([]string, 0)

	for _, client := range clients {
		for _, cidr := range strings.Split(client.Address, ",") {
			ip, err := util.GetIpFromCidr(strings.TrimSpace(cidr))
			if err != nil {
				log.WithFields(log.Fields{
					"err":  err,
					"cidr": err,
				}).Error("failed to ip from cidr")
			} else {
				reserverIps = append(reserverIps, ip)
			}
		}
	}

	for _, cidr := range strings.Split(server.Address, ",") {
		ip, err := util.GetIpFromCidr(strings.TrimSpace(cidr))
		if err != nil {
			log.WithFields(log.Fields{
				"err":  err,
				"cidr": err,
			}).Error("failed to ip from cidr")
		} else {
			reserverIps = append(reserverIps, ip)
		}
	}

	return reserverIps, nil
}
