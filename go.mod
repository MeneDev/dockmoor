module github.com/MeneDev/dockmoor

go 1.12

require (
	github.com/Microsoft/go-winio v0.4.13-0.20190408173621-84b4ab48a507 // indirect
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/bitly/go-hostpool v0.0.0-20171023180738-a3a6125de932 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/bugsnag/bugsnag-go v1.5.1 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/cloudflare/cfssl v0.0.0-20190506234652-e03d70fc14f2 // indirect
	github.com/containerd/continuity v0.0.0-20190426062206-aaeac12a7ffc // indirect
	github.com/docker/cli v0.0.0-20181026145426-51668a30f262
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/docker/docker-credential-helpers v0.6.0 // indirect
	github.com/docker/go v1.5.1-1 // indirect
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/google/certificate-transparency-go v1.0.21 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jessevdk/go-flags v1.4.0
	github.com/jinzhu/gorm v1.9.7 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/lib/pq v1.1.1 // indirect
	github.com/mattn/go-shellwords v1.0.5
	github.com/miekg/pkcs11 v1.0.2 // indirect
	github.com/moby/buildkit v0.3.3
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/runc v1.0.1-0.20190307181833-2b18fe1d885e // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.3 // indirect
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/testcontainers/testcontainers-go v0.0.5
	github.com/theupdateframework/notary v0.6.1 // indirect
	golang.org/x/sys v0.0.0-20190312061237-fead79001313 // indirect
	golang.org/x/text v0.3.1-0.20181227161524-e6919f6577db // indirect
	gopkg.in/dancannon/gorethink.v3 v3.0.5 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5 // indirect
)

// replace github.com/testcontainers/testcontainers-go => ../testcontainers-go
replace github.com/testcontainers/testcontainers-go => github.com/MeneDev/testcontainers-go v0.0.5
