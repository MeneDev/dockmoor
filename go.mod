module github.com/MeneDev/dockmoor

go 1.12

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/bitly/go-hostpool v0.0.0-20171023180738-a3a6125de932 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/cloudflare/cfssl v0.0.0-20190506234652-e03d70fc14f2 // indirect
	github.com/containerd/containerd v1.3.3 // indirect
	github.com/containerd/ttrpc v1.0.0 // indirect
	github.com/containerd/typeurl v0.0.0-20200205145503-b45ef1f1f737 // indirect
	github.com/coreos/go-systemd/v22 v22.0.0 // indirect
	github.com/docker/cli v0.0.0-20200210162036-a4bedce16568
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v0.7.3-0.20190506211059-b20a14b54661
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go v1.5.1-1 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/golang/protobuf v1.3.4 // indirect
	github.com/google/certificate-transparency-go v1.0.21 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/jessevdk/go-flags v1.4.0
	github.com/jinzhu/gorm v1.9.12 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-shellwords v1.0.10
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/miekg/pkcs11 v1.0.3 // indirect
	github.com/moby/buildkit v0.3.3
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/runtime-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/procfs v0.0.10 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.6 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	github.com/testcontainers/testcontainers-go v0.0.0-00010101000000-000000000000
	github.com/theupdateframework/notary v0.6.1 // indirect
	github.com/urfave/cli v1.22.2 // indirect
	github.com/xlab/handysort v0.0.0-20150421192137-fb3537ed64a1 // indirect
	golang.org/x/sys v0.0.0-20190830141801-acfa387b8d69 // indirect
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/dancannon/gorethink.v3 v3.0.5 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787 // indirect
)

// https://github.com/golang/go/issues/28489#issuecomment-528122909
// v19.03.6
replace github.com/docker/docker => github.com/docker/engine v0.0.0-20200204220554-5f6d6f3f2203

replace github.com/testcontainers/testcontainers-go => github.com/MeneDev/testcontainers-go v0.3.2-0.20200418134517-09f104aace0b

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190830141801-acfa387b8d69
