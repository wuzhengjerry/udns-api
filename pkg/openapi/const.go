package openapi

import "time"

const (
	HalfHourTTL    = 30 * time.Minute
	HourTTL        = 3 * time.Hour
	QueryWorkCount = 50
	OptWorkCount   = 10
	BasePermGroup  = "base_perm_group"
	TmpCookie      = "tapdsession=ee8922c6272c498b2743c59ca8417df0; t_u=f9ea5045e3a8fbb9%7C0725683f9c0ff94b; " +
		"t_uid=yingxiaosun; x-host-key-ngn=178d3756a01-e2c453ce3c8c0695f9709aa89832e31f2d2d9999; x-host-key-front=" +
		"178d3756b5d-727125e14488e27e4c33df34d12bb5dc6e605701; km_u=d96d89a37adbbbe057f7f981138bb392ed7d8d2bb32391" +
		"55e8180009fd52b35fdc8265c0e0922321; km_uid=yingxiaosun; x-host-key-oaback=178d4a2ae5d-dc567e34d7996cc235620" +
		"6ffc06bd24cfb57fe77; paas_perm_csrftoken=NqmQSfoB9Yqte5jl5jYlADXrTU3Od3BDPU8GgzQCJZ9Q4bsnvBQiLl4rikwva1RV; " +
		"paas_perm_sessionid=iqtihs9sj4r0ql0vsqyu4brmtkiwds2y; X-DEVOPS-PROJECT-ID=nmpc; x_host_key=178f9c16f3e-156b" +
		"5a20ea4190aa2ff3bb0bb2a2a2f22044b594; bk_uid=yingxiaosun; bk_ticket=dXEMsbw_JoEsHWjhX8kBd9TpSJbcB1QG9k414dT" +
		"l4ME; blueking_language=zh-cn; x-host-key-idcback=178fd2dc8c2-b00559e6c5e34d8dfb7b78ec8af162ff33b750ce; Digg" +
		"erTraceId=3dfed7d0-a407-11eb-b758-c1ab0c90c6f8; TCOA_TICKET=TOF4TeyJ2IjoiNCIsInRpZCI6ImRBaVpSN1RkTkhYVU53OFJ" +
		"OSnlDakpvaFdOV3FBeU5UIiwiaXNzIjoiMTAuMjguODMuMTQ4IiwiaGsiOiIiLCJpYXQiOiIyMDIxLTA0LTI1VDEwOjE2OjQ1LjYxOTgyOTY" +
		"4MyswODowMCIsImF1ZCI6IjAuMC4wLjAiLCJoYXNoIjoiMEEwQTAwQTE2NzdEN0E1QzZERTgzOUNCNzc4MDkwNzhBRjYzQUFGM0MzRTJEMjF" +
		"BQ0VFNDZDOTAzNDVFM0EzMSJ9; TCOA=dAiZR7TdNHXUNw8RNJyCjJohWNWqAyNT; RIO_TCOA_TICKET=tof:TOF4TeyJ2IjoiNCIsInRpZ" +
		"CI6ImRBaVpSN1RkTkhYVU53OFJOSnlDakpvaFdOV3FBeU5UIiwiaXNzIjoiMTAuMjguODMuMTQ4IiwiaGsiOiIiLCJpYXQiOiIyMDIxLTA0L" +
		"TI1VDEwOjE2OjQ1LjYxOTgyOTY4MyswODowMCIsImF1ZCI6IjAuMC4wLjAiLCJoYXNoIjoiMEEwQTAwQTE2NzdEN0E1QzZERTgzOUNCNzc4M" +
		"DkwNzhBRjYzQUFGM0MzRTJEMjFBQ0VFNDZDOTAzNDVFM0EzMSJ9"
)
