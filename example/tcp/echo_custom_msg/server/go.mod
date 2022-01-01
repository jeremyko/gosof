module tcpCustomMsgServer

go 1.16

replace github.com/jeremyko/gosof/example/tcp/echo_custom_msg/custom_msg_def => ../custom_msg_def/

require (
	github.com/jeremyko/gosof v0.0.0-20220101151901-97b6f40e9e78 // indirect
	github.com/jeremyko/gosof/example/tcp/echo_custom_msg/custom_msg_def v0.0.0-00010101000000-000000000000
)
