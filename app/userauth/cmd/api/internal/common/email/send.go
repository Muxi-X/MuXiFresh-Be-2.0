package email

import (
	code "MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/code"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"MuXiFresh-Be-2.0/common/convert"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/tool"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/jordan-wright/email"
	"log"
	"net/smtp"
	"net/textproto"
)

type SenderConf struct {
	Host     string
	Port     string
	UserName string
	Password string
}

var senderConf SenderConf

func Load(c config.Config) {
	copier.Copy(&senderConf, c.EmailConf)
}

type Msg struct {
	Email string `json:"email"`
	Type  string `json:"type"`
}

func (message *Msg) send() error {
	Email := message.Email
	Type := message.Type
	if Type != globalKey.Register && Type != globalKey.SetPassword && Type != globalKey.SetEmail {
		return fmt.Errorf("invalid email type")
	}
	//生成一个验证码
	randCode := tool.RandStringBytes(6)
	html := "<div style=\"position: relative;\">\n    <table width=\"720px\" align=\"center\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\">\n        <tbody>\n            <tr>\n                <td class=\"p-80 mpy-35 mpx-15\" bgcolor=\"#FFFFFF\" align=\"center\">\n                    <table width=\"720px\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\" align=\"center\">\n                        <tbody>\n                            <tr>\n                                <td style=\"width: 720px;\" align=\"center\">\n                                    <svg>\n                                        <path\n                                            d=\"M 50.764 34.952 C 44.869 36.821, 40.780 40.276, 38.392 45.404 C 34.157 54.499, 35.561 59.627, 45.655 71.927 C 49.695 76.851, 53 81.069, 53 81.300 C 53 81.531, 47.487 81.688, 40.750 81.647 C 30.061 81.583, 27.919 81.855, 23.940 83.784 C 14.489 88.365, 10.328 99.886, 14.908 108.790 C 18.779 116.314, 24.597 120, 32.605 120 C 38.285 120, 42.675 117.728, 51.742 110.096 C 55.725 106.743, 59.212 104, 59.492 104 C 59.771 104, 60 109.371, 60 115.935 C 60 122.499, 60.479 129.137, 61.064 130.685 C 63.494 137.117, 72.007 142.934, 79.054 142.978 C 81.008 142.990, 84.833 141.883, 87.554 140.517 C 97.375 135.587, 101.205 124.449, 96.381 114.845 C 95.647 113.385, 92.336 108.779, 89.023 104.610 C 85.711 100.441, 82.999 96.686, 82.999 96.265 C 82.998 95.844, 88.044 95.655, 94.213 95.844 C 100.678 96.043, 107.013 95.714, 109.169 95.068 C 122.684 91.018, 127.264 74.173, 117.626 63.962 C 113.451 59.539, 109.524 57.857, 103.371 57.857 C 97.146 57.857, 93.735 59.580, 82.500 68.401 L 75.500 73.897 75 60.534 C 74.519 47.671, 74.388 47.013, 71.500 42.925 C 66.841 36.330, 57.611 32.781, 50.764 34.952 M 158.525 73.186 C 158.142 74.277, 155.930 87.741, 153.463 104 C 153.171 105.925, 152.686 108.737, 152.386 110.250 C 151.875 112.827, 152.060 113, 155.318 113 L 158.795 113 159.972 106.250 C 160.620 102.537, 161.405 97.349, 161.718 94.721 C 162.030 92.092, 162.589 89.754, 162.959 89.525 C 163.330 89.296, 165.315 94.511, 167.371 101.114 L 171.110 113.118 174.264 112.809 C 177.371 112.505, 177.472 112.324, 181.115 100.500 C 183.149 93.900, 185.066 89.175, 185.376 90 C 185.686 90.825, 186.639 95.947, 187.493 101.383 C 188.347 106.818, 189.287 111.656, 189.582 112.133 C 190.237 113.193, 196.003 113.298, 195.991 112.250 C 195.967 110.077, 189.889 73.629, 189.426 72.881 C 189.127 72.396, 187.584 72, 185.997 72 C 183.185 72, 183.019 72.301, 179.493 83.750 C 177.503 90.213, 175.648 96.287, 175.370 97.250 C 175.093 98.213, 174.558 98.995, 174.183 98.989 C 173.807 98.984, 171.694 93.021, 169.486 85.739 L 165.471 72.500 162.229 72.186 C 160.214 71.991, 158.812 72.370, 158.525 73.186 M 200.360 73.505 C 200.042 74.333, 199.945 81.666, 200.143 89.802 C 200.550 106.500, 201.425 109.026, 207.705 111.650 C 212.952 113.842, 218.238 112.631, 221.969 108.383 L 225 104.930 225 88.402 L 225 71.873 221.750 72.187 L 218.500 72.500 218.222 87.431 C 218.056 96.361, 217.504 102.892, 216.850 103.681 C 215.130 105.753, 209.254 105.343, 208.035 103.066 C 207.458 101.986, 207 94.694, 207 86.566 L 207 72 203.969 72 C 202.192 72, 200.698 72.623, 200.360 73.505 M 232 72.700 C 232 73.086, 234.734 77.547, 238.075 82.614 L 244.150 91.827 237.575 101.959 C 233.959 107.532, 231 112.296, 231 112.546 C 231 112.796, 232.803 113, 235.007 113 C 238.879 113, 239.166 112.762, 243.568 105.894 C 246.073 101.986, 248.295 98.962, 248.507 99.174 C 248.719 99.386, 250.835 102.471, 253.209 106.029 C 257.095 111.855, 257.898 112.531, 261.263 112.810 C 263.767 113.018, 265 112.705, 265 111.860 C 265 111.167, 262.244 106.420, 258.876 101.312 L 252.752 92.025 258.876 82.454 C 262.244 77.189, 265 72.684, 265 72.441 C 265 72.199, 263.207 72, 261.015 72 C 257.199 72, 256.858 72.274, 252.926 78.500 C 250.668 82.075, 248.689 85, 248.528 85 C 248.367 85, 246.329 82.075, 244 78.500 C 239.989 72.343, 239.560 72, 235.883 72 C 233.747 72, 232 72.315, 232 72.700 M 272 92.500 L 272 113.127 275.250 112.813 L 278.500 112.500 278.500 92.500 L 278.500 72.500 275.250 72.187 L 272 71.873 272 92.500\"\n                                            stroke=\"none\" fill=\"#f6c148\" fill-rule=\"evenodd\"></path>\n                                        <path\n                                            d=\"M -0 99.503 L -0 199.006 151.250 198.753 L 302.500 198.500 302.755 99.250 L 303.010 0 151.505 0 L 0 0 -0 99.503 M 0.481 100 C 0.481 154.725, 0.602 177.112, 0.750 149.750 C 0.898 122.387, 0.898 77.612, 0.750 50.250 C 0.602 22.887, 0.481 45.275, 0.481 100 M 50.764 34.952 C 44.869 36.821, 40.780 40.276, 38.392 45.404 C 34.157 54.499, 35.561 59.627, 45.655 71.927 C 49.695 76.851, 53 81.069, 53 81.300 C 53 81.531, 47.487 81.688, 40.750 81.647 C 30.061 81.583, 27.919 81.855, 23.940 83.784 C 14.489 88.365, 10.328 99.886, 14.908 108.790 C 18.779 116.314, 24.597 120, 32.605 120 C 38.285 120, 42.675 117.728, 51.742 110.096 C 55.725 106.743, 59.212 104, 59.492 104 C 59.771 104, 60 109.371, 60 115.935 C 60 122.499, 60.479 129.137, 61.064 130.685 C 63.494 137.117, 72.007 142.934, 79.054 142.978 C 81.008 142.990, 84.833 141.883, 87.554 140.517 C 97.375 135.587, 101.205 124.449, 96.381 114.845 C 95.647 113.385, 92.336 108.779, 89.023 104.610 C 85.711 100.441, 82.999 96.686, 82.999 96.265 C 82.998 95.844, 88.044 95.655, 94.213 95.844 C 100.678 96.043, 107.013 95.714, 109.169 95.068 C 122.684 91.018, 127.264 74.173, 117.626 63.962 C 113.451 59.539, 109.524 57.857, 103.371 57.857 C 97.146 57.857, 93.735 59.580, 82.500 68.401 L 75.500 73.897 75 60.534 C 74.519 47.671, 74.388 47.013, 71.500 42.925 C 66.841 36.330, 57.611 32.781, 50.764 34.952 M 158.525 73.186 C 158.142 74.277, 155.930 87.741, 153.463 104 C 153.171 105.925, 152.686 108.737, 152.386 110.250 C 151.875 112.827, 152.060 113, 155.318 113 L 158.795 113 159.972 106.250 C 160.620 102.537, 161.405 97.349, 161.718 94.721 C 162.030 92.092, 162.589 89.754, 162.959 89.525 C 163.330 89.296, 165.315 94.511, 167.371 101.114 L 171.110 113.118 174.264 112.809 C 177.371 112.505, 177.472 112.324, 181.115 100.500 C 183.149 93.900, 185.066 89.175, 185.376 90 C 185.686 90.825, 186.639 95.947, 187.493 101.383 C 188.347 106.818, 189.287 111.656, 189.582 112.133 C 190.237 113.193, 196.003 113.298, 195.991 112.250 C 195.967 110.077, 189.889 73.629, 189.426 72.881 C 189.127 72.396, 187.584 72, 185.997 72 C 183.185 72, 183.019 72.301, 179.493 83.750 C 177.503 90.213, 175.648 96.287, 175.370 97.250 C 175.093 98.213, 174.558 98.995, 174.183 98.989 C 173.807 98.984, 171.694 93.021, 169.486 85.739 L 165.471 72.500 162.229 72.186 C 160.214 71.991, 158.812 72.370, 158.525 73.186 M 200.360 73.505 C 200.042 74.333, 199.945 81.666, 200.143 89.802 C 200.550 106.500, 201.425 109.026, 207.705 111.650 C 212.952 113.842, 218.238 112.631, 221.969 108.383 L 225 104.930 225 88.402 L 225 71.873 221.750 72.187 L 218.500 72.500 218.222 87.431 C 218.056 96.361, 217.504 102.892, 216.850 103.681 C 215.130 105.753, 209.254 105.343, 208.035 103.066 C 207.458 101.986, 207 94.694, 207 86.566 L 207 72 203.969 72 C 202.192 72, 200.698 72.623, 200.360 73.505 M 232 72.700 C 232 73.086, 234.734 77.547, 238.075 82.614 L 244.150 91.827 237.575 101.959 C 233.959 107.532, 231 112.296, 231 112.546 C 231 112.796, 232.803 113, 235.007 113 C 238.879 113, 239.166 112.762, 243.568 105.894 C 246.073 101.986, 248.295 98.962, 248.507 99.174 C 248.719 99.386, 250.835 102.471, 253.209 106.029 C 257.095 111.855, 257.898 112.531, 261.263 112.810 C 263.767 113.018, 265 112.705, 265 111.860 C 265 111.167, 262.244 106.420, 258.876 101.312 L 252.752 92.025 258.876 82.454 C 262.244 77.189, 265 72.684, 265 72.441 C 265 72.199, 263.207 72, 261.015 72 C 257.199 72, 256.858 72.274, 252.926 78.500 C 250.668 82.075, 248.689 85, 248.528 85 C 248.367 85, 246.329 82.075, 244 78.500 C 239.989 72.343, 239.560 72, 235.883 72 C 233.747 72, 232 72.315, 232 72.700 M 272 92.500 L 272 113.127 275.250 112.813 L 278.500 112.500 278.500 92.500 L 278.500 72.500 275.250 72.187 L 272 71.873 272 92.500\"\n                                            stroke=\"none\" fill=\"#ffffff\" fill-rule=\"evenodd\"></path>\n                                    </svg>\n                                </td>\n                            </tr>\n                            <tr>\n                                <td>\n                                    <table width=\"100%\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\">\n                                        <tbody>\n                                            <tr>\n                                                <td\n                                                    style=\"font-size:20px; line-height:42px; font-family:Arial, sans-serif, 'Motiva Sans'; text-align:left; padding-bottom: 30px; color:#002333; font-weight:bold;\">\n                                                    <span style=\"color: #F4AB4C;\">\n                                                        同学你好\n                                                    </span>\n                                                </td>\n                                            </tr>\n                                        </tbody>\n                                    </table>\n                                    <table width=\"100%\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\">\n                                        <tbody>\n                                            <tr>\n                                                <td class=\"text-18 c-grey4 pb-30\"\n                                                    style=\"vertical-align: bottom;font-size:18px; line-height:25px; font-family:Arial, sans-serif, 'Motiva Sans'; text-align:left; color:#002333; padding-bottom: 30px;\">\n                                                    您正在进行 验证操作，这是您验证帐户所需的验证码，有效时间为 10 分钟。\n                                                </td>\n                                            </tr>\n                                        </tbody>\n                                    </table>\n                                    <table width=\"100%\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\" align=\"center\">\n                                        <tbody>\n                                            <tr>\n                                                <td class=\"pb-70 mpb-50\" style=\"padding-bottom: 70px;\" align=\"center\">\n                                                    <table width=\"100%\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\"\n                                                        style=\"background-color: #fae8d2;\">\n                                                        <tbody>\n                                                            <tr>\n                                                                <td class=\"py-30 px-56\"\n                                                                    style=\"padding-top: 30px; padding-bottom: 30px; \">\n                                                                    <table width=\"720px\" border=\"0\" cellspacing=\"0\"\n                                                                        cellpadding=\"0\">\n                                                                        <tbody>\n                                                                            <tr>\n                                                                                <td\n                                                                                    style=\"font-size:18px; line-height:25px; font-family:Arial, sans-serif, 'Motiva Sans'; color:#8f98a0; text-align:center;\">\n                                                                                    验证码\n                                                                                </td>\n                                                                            </tr>\n                                                                            <tr>\n                                                                                <td style=\"padding-bottom: 16px\">\n                                                                                </td>\n                                                                            </tr>\n                                                                            <tr>\n                                                                                <td class=\"title-48 c-blue1 fw-b a-center\"\n                                                                                    style=\"font-size:48px; line-height:52px; font-family:Arial, sans-serif, 'Motiva Sans'; color:#f5a338; font-weight:bold; text-align:center;\">\n                                                                                    " + randCode + "\n                                                                                </td>\n                                                                            </tr>\n                                                                        </tbody>\n                                                                    </table>\n                                                                </td>\n                                                            </tr>\n                                                        </tbody>\n                                                    </table>\n                                                </td>\n                                            </tr>\n                                        </tbody>\n                                    </table>\n                                    <table width=\"100%\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\">\n                                    </table>\n                                    <table width=\"100%\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\">\n                                        <tbody>\n                                            <tr>\n                                                <td class=\"title-36 pb-30 c-grey6 fw-b\"\n                                                    style=\"font-size:30px; line-height:34px; font-family:Arial, sans-serif, 'Motiva Sans'; text-align:left; padding-bottom: 20px; color:#002333; font-weight:bold;\">\n                                                    不是您？\n                                                </td>\n                                            </tr>\n                                        </tbody>\n                                    </table>\n                                    <table width=\"100%\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\">\n                                        <tbody>\n                                            <tr>\n                                                <td class=\"text-18 c-grey4 pb-30\"\n                                                    style=\"font-size:18px; line-height:25px; font-family:Arial, sans-serif, 'Motiva Sans'; text-align:left; color:#002333; padding-bottom: 30px;\">\n                                                    如果这不是来自您的验证请求，请您忽略本邮件并\n                                                    <span style=\"color: #002333; font-weight: bold;\">\n                                                        不要将验证码转发给任何人\n                                                    </span>\n                                                    。\n                                                    <br>\n                                                    <br>\n                                                    此电子邮件包含一个代码，您需要用它验证您的帐户。切勿与任何人分享此代码。\n                                                </td>\n                                            </tr>\n                                        </tbody>\n                                    </table>\n                                    <table width=\"100%\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\">\n                                        <tbody>\n                                            <tr>\n                                                <td class=\"pt-30\" style=\"padding-top: 30px;\">\n                                                    <table width=\"100%\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\">\n                                                        <tbody>\n                                                            <tr>\n                                                                <td width=\"5px\" height=\"120px\"\n                                                                    style=\"background-color: #f5a338;text-align:left;\">\n                                                                </td>\n                                                                <td class=\"img\" width=\"37\" style=\"text-align:left;\">\n                                                                </td>\n                                                                <td>\n                                                                    <table width=\"100%\" border=\"0\" cellspacing=\"0\"\n                                                                        cellpadding=\"0\">\n                                                                        <tbody>\n                                                                            <tr>\n                                                                                <td class=\"text-16 py-20 c-grey4 fallback-font\"\n                                                                                    style=\"font-size:16px; line-height:22px; font-family:Arial, sans-serif, 'Motiva Sans'; text-align:left; padding-top: 20px; padding-bottom: 20px; color:#002333;\">\n                                                                                    祝您愉快，\n                                                                                    <br>\n                                                                                    木犀团队\n                                                                                </td>\n                                                                            </tr>\n                                                                        </tbody>\n                                                                    </table>\n                                                                </td>\n                                                            </tr>\n                                                        </tbody>\n                                                    </table>\n                                                </td>\n                                            </tr>\n                                        </tbody>\n                                    </table>\n                                </td>\n                            </tr>\n                        </tbody>\n                    </table>\n                </td>\n            </tr>\n            <tr>\n            </tr>\n        </tbody>\n    </table>\n</div>"
	subject := fmt.Sprintf("木犀团队招新系统 - %v", convert.TypeCvtChinese(Type))
	//发送
	e := &email.Email{
		To:      []string{Email},
		From:    senderConf.UserName,
		Subject: subject,
		HTML:    []byte(html),
		Headers: textproto.MIMEHeader{},
	}
	err := e.Send(senderConf.Host+":"+senderConf.Port, smtp.PlainAuth("", senderConf.UserName, senderConf.Password, senderConf.Host))
	if err != nil {
		return err
	}
	//存到redis
	err = code.SetEmailCode(Type, Email, randCode)
	return err
}

var queue chan *Msg

func Push(msg *Msg) error {
	select {
	case queue <- msg:
		return nil
	default:
		return fmt.Errorf("email queue push failed")
	}
}

func Sender() {
	queue = make(chan *Msg, 100)
	var msg *Msg
	var err error
	for {
		msg = <-queue
		err = msg.send()
		if err != nil {
			log.Print("emailSender: send failed")
		}
	}
}
