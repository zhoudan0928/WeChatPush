package main

import (
	"bestrui/wechatpush/mail"
	"bestrui/wechatpush/openwechat"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	// 设置日志输出
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式

	// 注册消息处理函数
	bot.MessageHandler = handleMessage

	// 创建热存储容器对象
	reloadStorage := openwechat.NewFileHotReloadStorage("/app/data/storage.json")
	defer reloadStorage.Close()

	// 登录
	if err := bot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		log.Printf("热登录失败: %v", err)
		log.Println("尝试正常登录")
		if err := bot.Login(); err != nil {
			log.Fatalf("登录失败: %v", err)
		}
	}

	log.Println("登录成功")

	// 阻塞主程序
	bot.Block()
}

func handleMessage(msg *openwechat.Message) {
	if msg.IsSendBySelf() {
		return
	}

	var sender string
	var content string

	if msg.IsSendByFriend() {
		friendSender, err := msg.Sender()
		if err != nil {
			log.Printf("获取发送者信息失败: %v", err)
			return
		}
		sender = friendSender.RemarkName
		if sender == "" {
			sender = friendSender.NickName
		}
	} else {
		groupSender, err := msg.SenderInGroup()
		if err != nil {
			log.Printf("获取群聊发送者信息失败: %v", err)
			return
		}
		sender = groupSender.NickName
	}

	switch {
	case msg.IsText():
		content = msg.Content
	case msg.IsPicture():
		content = "[图片]"
	case msg.IsVoice():
		content = "[语音]"
	case msg.IsVideo():
		content = "[视频]"
	case msg.IsEmoticon():
		content = "[动画表情]"
	default:
		content = "[未知类型消息]"
	}

	log.Printf("%s: %s", sender, content)

	if !msg.IsSendByGroup() || (msg.IsText() && strings.Contains(msg.Content, "@所有人")) {
		for i := 0; i < 3; i++ { // 重试3次
			if err := mail.SendEmail(sender, content); err != nil {
				log.Printf("发送邮件失败 (尝试 %d/3): %v", i+1, err)
				time.Sleep(time.Second * 2) // 等待2秒后重试
			} else {
				log.Printf("邮件发送成功: %s - %s", sender, content)
				return
			}
		}
		log.Printf("发送邮件失败，已达到最大重试次数")
	}
}
