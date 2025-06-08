# Telegram interactive bot 

A golang version of [Telegram-interactive-bot][tgbot]


## I. Introduction

An open-source bidirectional bot for Telegram. It helps to avoid spam messages and allows restricted clients to contact you smoothly.

[中文文档][cnmd] 


### Features

- When a client contacts customer service through the bot, all messages will be completely forwarded to the background management group, creating a separate sub-forum named after the client's information to distinguish them from other clients.
- Replies from customer service in the sub-forum can be directly sent to the client.
- Customer service can configure whether to continue the conversation with the client by closing/opening the sub-forum.
- Provides a permanent ban solution. There is a switch in the environment variables.
- Provides a /clear command to clear all messages in the sub-forum, also deleting user messages (not recommended, but sometimes necessary). There is a switch in the environment variables.

### Advantages

- By using sub-forums, multiple management members can be added to share the customer service workload.
- Complete communication records with clients can be intuitively retained.
- It's possible to know which customer service representative replied to a particular message, maintaining coherent customer service.

## 2. Preparation

The main principle of this bot is to forward the conversation between the client and the bot to a group (preferably a private group) and categorize each client's messages into a sub-category. Therefore, before starting, you need to:

1. Find @BotFather and apply for a bot.
2. Obtain the bot's token.
3. Create a group (set as public as needed).
4. Enable "Topics" in the group.
5. Add your bot to the group and promote it to an administrator.
6. Remember to include "Message Management" and "Topic Management" in the administrative permissions.
7. Use the bot @GetTheirIDBot to obtain the built-in ID of the group and the user ID of the administrator.
8. Use the bot @GetTheirIDBot to get the built-in ID and administrator user ID of the group.


## 3. Deployment and Execution



# About

Inspired by [Telegram-interactive-bot][tgbot]

[tgbot]: https://github.com/MiHaKun/Telegram-interactive-bot
[cnmd]: https://github.com/Guaderxx/interbot/blob/main/README.md

