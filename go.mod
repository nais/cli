module github.com/nais/cli

go 1.24.2

tool (
	github.com/onsi/ginkgo/v2/ginkgo
	golang.org/x/tools/cmd/deadcode
	golang.org/x/vuln/cmd/govulncheck
	honnef.co/go/tools/cmd/staticcheck
	mvdan.cc/gofumpt
)

require (
	cloud.google.com/go/cloudsqlconn v1.16.1
	github.com/GoogleCloudPlatform/cloudsql-proxy v1.37.7
	github.com/go-logr/logr v1.4.2
	github.com/goccy/go-yaml v1.17.1
	github.com/lib/pq v1.10.9
	github.com/mailgun/raymond/v2 v2.0.48
	github.com/mitchellh/go-ps v1.0.0
	github.com/nais/device v0.0.0-20231030073017-240e8a0234d9
	github.com/nais/liberator v0.0.0-20250210140717-70bb7da67963
	github.com/onsi/ginkgo/v2 v2.23.4
	github.com/onsi/gomega v1.37.0
	github.com/pterm/pterm v0.12.80
	github.com/sethvargo/go-retry v0.3.0
	github.com/stretchr/testify v1.10.0
	github.com/urfave/cli/v2 v2.27.6
	github.com/xeipuuv/gojsonschema v1.2.0
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.35.0
	go.opentelemetry.io/otel/metric v1.35.0
	go.opentelemetry.io/otel/sdk v1.35.0
	go.opentelemetry.io/otel/sdk/metric v1.35.0
	golang.org/x/oauth2 v0.29.0
	golang.org/x/sync v0.13.0
	google.golang.org/api v0.229.0
	google.golang.org/grpc v1.71.1
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.33.0
	k8s.io/apimachinery v0.33.0
	k8s.io/client-go v0.33.0
	k8s.io/klog/v2 v2.130.1
	k8s.io/utils v0.0.0-20241210054802-24370beab758
	sigs.k8s.io/controller-runtime v0.20.4
)

require (
	atomicgo.dev/cursor v0.2.0 // indirect
	atomicgo.dev/keyboard v0.2.9 // indirect
	atomicgo.dev/schedule v0.1.0 // indirect
	cloud.google.com/go/auth v0.16.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.12.1 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/gookit/color v1.5.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.1 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/lithammer/fuzzysearch v1.1.8 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/exp v0.0.0-20241009180824-f66d83c29e7c // indirect
	golang.org/x/exp/typeparams v0.0.0-20241009180824-f66d83c29e7c // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/telemetry v0.0.0-20241004145657-5eebfecbdf1f // indirect
	golang.org/x/term v0.31.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	golang.org/x/tools v0.32.0 // indirect
	golang.org/x/vuln v1.1.4 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250414145226-207652e42e2e // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	honnef.co/go/tools v0.6.1 // indirect
	k8s.io/apiextensions-apiserver v0.32.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff // indirect
	mvdan.cc/gofumpt v0.8.0 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.6.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

retract [v0.0.2, v0.0.1]
