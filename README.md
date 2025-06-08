# Telegram Interactive Bot

A golang version of [Telegram-interactive-bot][tgbot]


## 一、简介

Telegram的开源双向机器人。避免垃圾信息；让被限制的客户可以顺利联系到你。 支持后台多客服。在后台群组，可以安排多个客服以同一个机器人身份持续和客户沟通。

[English][enmd]


### 特色

- 当客户通过机器人联系客服时，所有消息将被完整转发到后台管理群组，生成一个独立的以客户信息命名子论坛，用来和其他客户区分开来。
- 客服在子论坛中的回复，可以直接回复给客户。
- 客服可以通过关闭/开启子论坛来配置是否继续和客户对话。
- 提供永久封禁方案。配置文件内有开关。
- 提供 /clear 命令，可以清除子论坛内的所有消息，同时也删除用户消息（极其不推荐如此使用，不过奈何也确实有时候有必要）。配置文件内有开关。

### 优势

- 借助子论坛，可以增加多个管理成员，分担客服压力。
- 可以直观的保留和客户沟通的完整通讯记录。
- 可以得知某句话是哪个客服回复的，维系连贯的客户服务。


## 二、准备工作

本机器人的主要原理是将客户和机器人的对话，转发到一个群内（自用，最好是私有群），并归纳每个客户的消息到一个子版块。 所以，在开工前，你需要：

1. 找 @BotFather 申请一个机器人。
2. 获取机器人的token
3. 建立一个群组（按需设置是否公开）
4. 群组的“话题功能”打开。
5. 将自己的机器人，拉入群组。提升权限为管理员。
6. 管理权限切记包含消息管理，话题管理。
7. 通过机器人 @GetTheirIDBot 获取群组的内置ID和管理员用户ID。


## 三、部署运行

### go install

1. 下载可执行文件：`go install github.com/Guaderxx/interbot/amd/app1`
2. 复制 `config/example.toml` 配置文件，修改其中参数，其中 `bot_token, mongouri, mdb, admin_group_id, admin_user_ids` 为必填项。
3. 启动：`app1 --config locale_config.toml`

# 关于

本项目灵感来源于 [Telegram-interactive-bot][tgbot]

[tgbot]: https://github.com/MiHaKun/Telegram-interactive-bot
[enmd]: https://github.com/Guaderxx/interbot/blob/main/README_en.md
