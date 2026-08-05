// Package controller Repository 访问辅助函数
// 所有 controller 文件通过此文件共享 repository 访问器
package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/di"
	"github.com/MiLab-Bit/OpenFastToken/repository"
)

func userRepo() repository.UserRepository         { return di.Default().User }
func tokenRepo() repository.TokenRepository       { return di.Default().Token }
func channelRepo() repository.ChannelRepository   { return di.Default().Channel }
func skillRepo() repository.SkillRepository       { return di.Default().Skill }
