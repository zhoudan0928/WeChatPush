package main

import (
	"bestrui/wechatpush/mail"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eatmoreapple/openwechat"
)

func main() {
	// 设置日志输出
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 创建热存储容器对象
	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")
	defer reloadStorage.Close()

	// 创建一个新的机器人实例
	bot := openwechat.DefaultBot(openwechat.Desktop)

	// 注册消息处理函数
	bot.MessageHandler = handleMessage

	// 注册登录事件
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// 登录
	if err := bot.Login(); err != nil {
		log.Printf("登录失败: %v", err)
		return
	}

	// 获取登陆的用户
	self, err := bot.GetCurrentUser()
	if err != nil {
		log.Printf("获取当前用户失败: %v", err)
		return
	}
	log.Printf("登录成功: %s", self.NickName)

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
	} else if msg.IsSendByGroup() {
		groupSender, err := msg.SenderInGroup()
		if err != nil {
			log.Printf("获取群聊发送者信息失败: %v", err)
			return
		}
		sender = groupSender.NickName
	} else {
		log.Println("未知的消息发送者类型")
		return
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
