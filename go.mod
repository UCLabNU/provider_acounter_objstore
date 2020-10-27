module acounter_objstore

go 1.15

require (
	github.com/UCLabNU/proto_pflow v0.0.1
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/shirou/gopsutil v2.20.9+incompatible // indirect
	github.com/synerex/proto_pcounter v0.0.6
	github.com/synerex/proto_storage v0.2.0
	github.com/synerex/synerex_api v0.4.2
	github.com/synerex/synerex_proto v0.1.9
	github.com/synerex/synerex_sxutil v0.6.1
	golang.org/x/net v0.0.0-20201016165138-7b1cca2348c0 // indirect
	golang.org/x/sys v0.0.0-20201018230417-eeed37f84f13 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20201019141844-1ed22bb0c154 // indirect
	google.golang.org/grpc v1.33.0 // indirect
	google.golang.org/protobuf v1.25.0
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace github.com/synerex/proto_pcounter v0.0.6 => github.com/nagata-yoshiteru/proto_pcounter v0.0.10

replace github.com/synerex/synerex_proto v0.1.9 => github.com/nagata-yoshiteru/synerex_proto v0.1.10
