package provision_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/mainflux/mainflux/provision"
	"github.com/mainflux/mainflux/provision/mocks"
	SDK "github.com/mainflux/mainflux/sdk/go"

	logger "github.com/mainflux/mainflux/logger"
	"github.com/stretchr/testify/assert"
)

var (
	cfg = provision.Config{
		Bootstrap: provision.Bootstrap{
			AutoWhiteList: true,
			Provision:     true,
			Content:       "",
			X509Provision: true,
		},
		Server: provision.ServiceConf{
			MfPass: "test",
			MfUser: "test@example.com",
		},
	}
	log, _ = logger.New(os.Stdout, "info")
)

func TestProvision(t *testing.T) {
	// Create multiple services with different configurations.
	certs := mocks.NewCertsSDK()
	sdk := mocks.NewSDK()
	svc := provision.New(cfg, sdk, certs, log)

	cases := []struct {
		desc        string
		externalID  string
		externalKey string
		svc         provision.Service
		err         error
	}{
		{
			desc:        "Provision successfully",
			externalID:  "id",
			externalKey: "key",
			svc:         svc,
			err:         nil,
		},
		{
			desc:        "Provision already existing config",
			externalID:  "id",
			externalKey: "key",
			svc:         svc,
			err:         SDK.ErrFailedCreation,
		},
	}

	for _, tc := range cases {
		_, err := tc.svc.Provision("", tc.externalID, tc.externalKey)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}

}