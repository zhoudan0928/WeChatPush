package main

import (
	"bestrui/wechatpush/mail"
	"bestrui/wechatpush/openwechat"
	"log"
	"strings"
)

func main() {
	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式

	// 创建热存储容器对象
	reloadStorage := openwechat.NewFileHotReloadStorage("/app/data/storage.json")

	defer reloadStorage.Close()

	// 登录
	if err := bot.HotLogin(reloadStorage); err != nil {
		log.Println("热登陆失败，尝试免扫码登录")
		if err := bot.PushLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
			log.Fatalf("登录失败: %v", err)
		}
	}

	log.Println("登陆成功")

	bot.MessageHandler = func(msg *openwechat.Message) {
		if msg.IsSendBySelf() { //自己发送的消息
			//跳过
			return
		} else if msg.IsSendByFriend() { //好友发送的消息
			friendSender, err := msg.Sender()
			if err != nil {
				log.Printf("获取发送者信息失败: %v", err)
				return
			}

			friendSenderName := friendSender.RemarkName
			if len(friendSender.RemarkName) == 0 {
				friendSenderName = friendSender.NickName
			}

			var content string
			if msg.IsText() {
				content = msg.Content
				log.Printf("%s: %s", friendSenderName, content)
			} else if msg.IsPicture() {
				content = "[图片]"
				log.Printf("%s: [图片]", friendSenderName)
			} else if msg.IsVoice() {
				content = "[语音]"
				log.Printf("%s: [语音]", friendSenderName)
			} else if msg.IsVideo() {
				content = "[视频]"
				log.Printf("%s: [视频]", friendSenderName)
			} else if msg.IsEmoticon() {
				content = "[动画表情]"
				log.Printf("%s: [动画表情]", friendSenderName)
			} else {
				content = "[未知类型消息]"
				log.Printf("%s: [未知类型消息]", friendSenderName)
			}

			if err := mail.SendEmail(friendSenderName, content); err != nil {
				log.Printf("发送邮件失败: %v", err)
			}
		} else { //群聊发送的消息
			groupSender, err := msg.SenderInGroup()
			if err != nil {
				log.Printf("获取群聊发送者信息失败: %v", err)
				return
			}
			if msg.IsText() {
				//群聊中只接受 @所有人 消息
				if strings.Contains(msg.Content, "@所有人") {
					log.Printf("%s: %s", groupSender.NickName, msg.Content)
					if err := mail.SendEmail(groupSender.NickName, msg.Content); err != nil {
						log.Printf("发送邮件失败: %v", err)
					}
				}
			}
		}
	}

	bot.Block()
}
