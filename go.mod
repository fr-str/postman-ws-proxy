module postman-proxy

go 1.19

require (
	github.com/gorilla/websocket v1.5.0
	github.com/main-kube/util v0.0.0-20220824130840-1ae10d265801
	github.com/rs/zerolog v1.27.0
)

require (
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/tidwall/gjson v1.14.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/wI2L/jsondiff v0.2.0 // indirect
	golang.org/x/exp v0.0.0-20220713135740-79cabaa25d75 // indirect
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654 // indirect
	syslabit.com/git/syslabit/log v0.0.0-20210815142336-c50bef1ee7dd // indirect
)

replace github.com/main-kube/util/safe => ../util/safe
