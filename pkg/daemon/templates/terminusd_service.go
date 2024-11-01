package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var (
	// TerminusdService defines the template of terminusd's service for systemd.
	TerminusdService = template.Must(template.New("terminusd.service").Parse(
		dedent.Dedent(`[Unit]
Description=terminusd
After=network.target
StartLimitIntervalSec=0

[Service]
User=root
EnvironmentFile=/etc/systemd/system/terminusd.service.env
ExecStart=/usr/local/bin/terminusd
RestartSec=10s
LimitNOFILE=40000
Restart=always

[Install]
WantedBy=multi-user.target
    `)))
)
