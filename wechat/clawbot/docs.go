// Package clawbot 提供微信 iLink Bot 的核心交互工具函数。
//
// 本包是无状态的工具库，封装 QR 登录、消息轮询、消息回复等操作。
// 不做任何持久化；调用方自行负责凭据存储、消息路由和 AI 接入等上层逻辑。
package clawbot
