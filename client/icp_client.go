package client

/*
	备案查询:
	工信部 https://beian.miit.gov.cn/ 只能网页、小程序
	阿里云
	腾讯云
*/

type ICPClient struct {
}

func NewICPClient() *ICPClient {
	return &ICPClient{}
}
