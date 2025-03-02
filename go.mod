module github.com/projectcalico/calico

go 1.23.2

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/Microsoft/hcsshim v0.11.6
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/aws/aws-sdk-go-v2 v1.25.0
	github.com/aws/aws-sdk-go-v2/config v1.27.0
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.15.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.148.0
	github.com/aws/smithy-go v1.20.0
	github.com/bits-and-blooms/bitset v1.13.0
	github.com/buger/jsonparser v1.1.1
	github.com/container-storage-interface/spec v1.9.0
	github.com/containernetworking/cni v1.2.0
	github.com/containernetworking/plugins v1.1.1
	github.com/coreos/go-semver v0.3.1
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/docker/distribution v2.8.2+incompatible
	github.com/docker/docker v27.1.1+incompatible
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/envoyproxy/go-control-plane v0.12.0
	github.com/fsnotify/fsnotify v1.7.0
	github.com/go-ini/ini v1.67.0
	github.com/gofrs/flock v0.8.1
	github.com/gogo/googleapis v1.4.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.4
	github.com/golang/snappy v0.0.4
	github.com/google/btree v1.1.2
	github.com/google/go-cmp v0.6.0
	github.com/google/go-github/v53 v53.2.0
	github.com/google/gopacket v1.1.19
	github.com/google/netstack v0.0.0-20191123085552-55fcc16cd0eb
	github.com/google/safetext v0.0.0-20230106111101-7156a760e523
	github.com/google/uuid v1.6.0
	github.com/gruntwork-io/terratest v0.46.7
	github.com/ishidawataru/sctp v0.0.0-20230406120618-7ff4192f6ff2
	github.com/joho/godotenv v1.5.1
	github.com/json-iterator/go v1.1.12
	github.com/juju/clock v1.0.3
	github.com/juju/errors v1.0.0
	github.com/juju/mutex v0.0.0-20180619145857-d21b13acf4bf
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kelseyhightower/memkv v0.1.1
	github.com/libp2p/go-reuseport v0.4.0
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/mipearson/rfw v0.0.0-20170619235010-6f0a6f3266ba
	github.com/natefinch/atomic v1.0.1
	github.com/nmrshll/go-cp v0.0.0-20180115193924-61436d3b7cfa
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/ginkgo/v2 v2.17.1
	github.com/onsi/gomega v1.32.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/projectcalico/api v0.0.0-00010101000000-000000000000
	github.com/projectcalico/go-json v0.0.0-20161128004156-6219dc7339ba
	github.com/projectcalico/go-yaml-wrapper v0.0.0-20191112210931-090425220c54
	github.com/prometheus/client_golang v1.18.0
	github.com/prometheus/client_model v0.6.0
	github.com/prometheus/common v0.47.0
	github.com/prometheus/procfs v0.12.0
	github.com/safchain/ethtool v0.0.0-20210803160452-9aa261dae9b1
	github.com/shirou/gopsutil/v4 v4.24.11
	github.com/sirupsen/logrus v1.9.3
	github.com/skeema/knownhosts v1.3.0
	github.com/slack-go/slack v0.14.0
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.9.0
	github.com/tchap/go-patricia/v2 v2.3.1
	github.com/termie/go-shutil v0.0.0-20140729215957-bcacb06fecae
	github.com/urfave/cli/v2 v2.27.3
	github.com/vishvananda/netlink v1.2.1-beta.2.0.20240703200800-b54f85093f4a
	go.etcd.io/etcd/api/v3 v3.5.12
	go.etcd.io/etcd/client/pkg/v3 v3.5.12
	go.etcd.io/etcd/client/v2 v2.305.12
	go.etcd.io/etcd/client/v3 v3.5.12
	golang.org/x/crypto v0.24.0
	golang.org/x/net v0.26.0
	golang.org/x/sync v0.7.0
	golang.org/x/sys v0.26.0
	golang.org/x/text v0.16.0
	golang.org/x/time v0.5.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200324154536-ceff61240acf
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240701130421-f6361c86f094
	google.golang.org/grpc v1.65.0
	gopkg.in/go-playground/validator.v9 v9.30.2
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.30.5
	k8s.io/apiextensions-apiserver v0.30.5
	k8s.io/apimachinery v0.30.5
	k8s.io/apiserver v0.30.5
	k8s.io/client-go v0.30.5
	k8s.io/component-base v0.30.5
	k8s.io/klog/v2 v2.120.1
	k8s.io/kube-aggregator v0.30.5
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340
	k8s.io/kubernetes v1.30.5
	k8s.io/utils v0.0.0-20240310230437-4693a0247e57
	modernc.org/memory v1.7.2
	sigs.k8s.io/controller-runtime v0.17.0
	sigs.k8s.io/kind v0.22.0
	sigs.k8s.io/knftables v0.0.18
	sigs.k8s.io/network-policy-api v0.1.5
	sigs.k8s.io/yaml v1.4.0
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/GoogleCloudPlatform/k8s-cloud-provider v1.18.1-0.20220218231025-f11817397a1b // indirect
	github.com/JeffAshton/win_pdh v0.0.0-20161109143554-76bb4ee9f0ab // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230217124315-7d5c6f04bbb8 // indirect
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/alexflint/go-filemutex v1.1.0 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230305170008-8188dc5388df // indirect
	github.com/armon/circbuf v0.0.0-20150827004946-bbbad097214e // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/aws/aws-sdk-go v1.44.122 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.19.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.22.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.27.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/checkpoint-restore/go-criu/v5 v5.3.0 // indirect
	github.com/cilium/ebpf v0.9.1 // indirect
	github.com/cloudflare/circl v1.3.9 // indirect
	github.com/cncf/xds/go v0.0.0-20240423153145-555b57ec207b // indirect
	github.com/containerd/cgroups v1.1.0 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/containerd/containerd v1.6.26 // indirect
	github.com/containerd/ttrpc v1.2.2 // indirect
	github.com/coreos/go-iptables v0.6.0 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/ebitengine/purego v0.8.1 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.0.4 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.8.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-playground/locales v0.12.1 // indirect
	github.com/go-playground/universal-translator v0.0.0-20170327191703-71201497bace // indirect
	github.com/go-sql-driver/mysql v1.4.1 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/cadvisor v0.49.0 // indirect
	github.com/google/cel-go v0.17.8 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20211214055906-6f57359322fd // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/gruntwork-io/go-commons v0.8.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/leodido/go-urn v0.0.0-20181204092800-a67a23e1c1af // indirect
	github.com/libopenstorage/openstorage v1.0.0 // indirect
	github.com/lithammer/dedent v1.1.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mattn/go-zglob v0.0.2-0.20190814121620-e3c945676326 // indirect
	github.com/mdlayher/genetlink v1.0.0 // indirect
	github.com/mdlayher/netlink v1.1.0 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170603005431-491d3605edfb // indirect
	github.com/mrunalp/fileutils v0.5.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/opencontainers/runc v1.1.14 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20220909204839-494a5a6aca78 // indirect
	github.com/opencontainers/selinux v1.11.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/pquerna/otp v1.2.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/seccomp/libseccomp-golang v0.10.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/urfave/cli v1.22.4 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful v0.46.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.49.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.52.0 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.21.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/sdk v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/oauth2 v0.20.0 // indirect
	golang.org/x/term v0.21.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	golang.zx2c4.com/wireguard v0.0.20200121 // indirect
	google.golang.org/api v0.169.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20240213162025-012b6fc9bca9 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240701130421-f6361c86f094 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/gcfg.v1 v1.2.3 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/cloud-provider v0.30.5 // indirect
	k8s.io/component-helpers v0.30.5 // indirect
	k8s.io/controller-manager v0.30.5 // indirect
	k8s.io/cri-api v0.30.5 // indirect
	k8s.io/csi-translation-lib v0.30.5 // indirect
	k8s.io/dynamic-resource-allocation v0.30.5 // indirect
	k8s.io/kms v0.30.5 // indirect
	k8s.io/kube-scheduler v0.30.5 // indirect
	k8s.io/kubectl v0.30.5 // indirect
	k8s.io/kubelet v0.30.5 // indirect
	k8s.io/legacy-cloud-providers v0.30.5 // indirect
	k8s.io/mount-utils v0.30.5 // indirect
	k8s.io/pod-security-admission v0.30.5 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.29.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)

replace (
	github.com/projectcalico/api => ./api

	k8s.io/api => k8s.io/api v0.30.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.30.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.30.5
	k8s.io/apiserver => k8s.io/apiserver v0.30.5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.30.5
	k8s.io/client-go => k8s.io/client-go v0.30.5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.30.5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.30.5
	k8s.io/code-generator => k8s.io/code-generator v0.30.5
	k8s.io/component-base => k8s.io/component-base v0.30.5
	k8s.io/component-helpers => k8s.io/component-helpers v0.30.5
	k8s.io/controller-manager => k8s.io/controller-manager v0.30.5
	k8s.io/cri-api => k8s.io/cri-api v0.30.5
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.30.5
	k8s.io/endpointslice => k8s.io/endpointslice v0.30.5
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.30.5
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.30.5
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.30.5
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.30.5
	k8s.io/kubectl => k8s.io/kubectl v0.30.5
	k8s.io/kubelet => k8s.io/kubelet v0.30.5

	// Need replacements for all the k8s subsidiary projects that are pulled in indirectly because
	// the kubernets repo pulls them in via a replacement to its own vendored copies, which doesn't work for
	// transient imports.
	k8s.io/kubernetes => k8s.io/kubernetes v1.30.5
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.30.5
	k8s.io/metrics => k8s.io/metrics v0.30.5
	k8s.io/mount-utils => k8s.io/mount-utils v0.30.5
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.30.5
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.30.5
)
