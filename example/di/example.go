package main

import (
	"fmt"

	"github.com/kruily/gofastcrud/core/di"
)

// IUserService 用户服务接口
type IUserService interface {
	GetUser(id int) string
}

// UserService 用户服务实现
type UserService struct {
	// 依赖注入
	Repository *UserRepository `inject:""`
}

func (s *UserService) GetUser(id int) string {
	return s.Repository.FindUser(id)
}

// UserRepository 用户仓储
type UserRepository struct{}

func (r *UserRepository) FindUser(id int) string {
	return fmt.Sprintf("User %d", id)
}

func main() {
	// 创建容器
	container := di.New()

	// 注册依赖
	container.BindSingleton((*UserRepository)(nil), &UserRepository{})
	container.BindSingleton((*IUserService)(nil), &UserService{})

	// 解析服务
	var userService IUserService
	container.MustResolve(&userService)

	// 使用服务
	result := userService.GetUser(1)
	fmt.Println(result) // 输出: User 1
}
