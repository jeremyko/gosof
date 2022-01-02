module tcpCustomMsgClient

go 1.16

replace github.com/jeremyko/gosof/example/tcp/echo_custom_msg/custom_msg_def => ../custom_msg_def/

require (
	github.com/jeremyko/gosof v1.0.0
	github.com/jeremyko/gosof/example/tcp/echo_custom_msg/custom_msg_def v0.0.0-00010101000000-000000000000
)
